package postgresql

import (
	"database/sql"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type postgresqlSyncRepository struct {
	db *sql.DB
}

// NewPostgreSQLSyncRepository creates a new PostgreSQL implementation of the SyncRepository
func NewPostgreSQLSyncRepository(db *sql.DB) repository.SyncRepository {
	return &postgresqlSyncRepository{
		db: db,
	}
}

// UpdateAndDeleteData processes the sync data by creating, updating, and deleting entries
func (repo *postgresqlSyncRepository) UpdateAndDeleteData(data model.SyncData) error {
	// Begin a transaction
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Process time entries to be created
	for _, entry := range data.TimeEntriesToBeCreated {
		err = repo.createTimeEntry(tx, &entry)
		if err != nil {
			return err
		}
	}

	// Process time entries to be updated
	for _, entry := range data.TimeEntriesToBeUpdated {
		err = repo.updateTimeEntry(tx, &entry)
		if err != nil {
			return err
		}
	}

	// Process time entries to be deleted
	for _, entry := range data.TimeEntriesToBeDeleted {
		err = repo.deleteTimeEntry(tx, &entry)
		if err != nil {
			return err
		}
	}

	// Process projects to be created
	for _, project := range data.ProjectsToBeCreated {
		err = repo.createProject(tx, &project)
		if err != nil {
			return err
		}
	}

	// Process projects to be updated
	for _, project := range data.ProjectsToBeUpdated {
		err = repo.updateProject(tx, &project)
		if err != nil {
			return err
		}
	}

	// Process projects to be deleted
	for _, project := range data.ProjectsToBeDeleted {
		err = repo.deleteProject(tx, &project)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// GetUpdatedTimeEntriesOfUser retrieves time entries that have been updated since a specific changelog entry ID
func (repo *postgresqlSyncRepository) GetUpdatedTimeEntriesOfUser(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error) {
	query := `
		SELECT te.id, te.user_id, te.project_id, te.start_time, te.end_time, te.description, te.deleted,
			   p.id as project_id, p.name as project_name, p.user_id as project_user_id,
			   p.team_id as project_team_id, p.color as project_color, p.deadline::date as project_deadline,
			   p.hourly_rate as project_hourly_rate, p.time_budget as project_time_budget,
			   p.is_active as project_is_active, p.deleted as project_deleted,
			   cl.operation, cl.id
		FROM change_log cl
		LEFT JOIN time_entries te ON cl.entity_id = te.id
		LEFT JOIN projects p ON te.project_id = p.id
		WHERE cl.id > $1
		AND cl.entity_type = 'TimeEntry'
		AND (te.user_id = $2 OR te.id IS NULL)
		AND (cl.changed_by_client != $3 OR cl.changed_by_client IS NULL OR cl.changed_by_client = '')
		ORDER BY cl.id ASC
	`

	rows, err := repo.db.Query(query, sinceTimeLogEntry, userId, excludeClientId)
	if err != nil {
		return model.TimeEntrySyncResult{}, err
	}
	defer rows.Close()

	// Maps to track which entries to include in each category
	createdEntries := make(map[uuid.UUID]model.TimeEntry)
	updatedEntries := make(map[uuid.UUID]model.TimeEntry)
	deletedEntries := make(map[uuid.UUID]model.TimeEntry)

	for rows.Next() {
		var entry model.TimeEntry
		var project model.Project
		var operation string
		var teamID sql.NullString
		var deadline sql.NullTime
		var hourlyRate sql.NullString
		var timeBudget sql.NullInt64
		var isActive sql.NullBool
		var color sql.NullString
		var projectDeleted sql.NullBool
		var changelogID int64

		err := rows.Scan(
			&entry.ID, &entry.UserId, &entry.ProjectId, &entry.StartTime, &entry.EndTime, &entry.Description, &entry.Deleted,
			&project.ID, &project.Name, &project.UserId, &teamID, &color, &deadline,
			&hourlyRate, &timeBudget, &isActive, &projectDeleted,
			&operation, &changelogID,
		)
		if err != nil {
			return model.TimeEntrySyncResult{}, err
		}

		// Set project fields
		if teamID.Valid {
			uuid, _ := uuid.FromString(teamID.String)
			project.TeamID = &uuid
		}
		if color.Valid {
			project.Color = color.String
		} else {
			project.Color = "#1E90FF" // Default blue color
		}
		if deadline.Valid {
			project.Deadline = model.NewDateOnly(deadline.Time)
		}
		if hourlyRate.Valid {
			project.HourlyRate, _ = decimal.NewFromString(hourlyRate.String)
		}
		if timeBudget.Valid {
			project.TimeBudget = int(timeBudget.Int64)
		}
		if isActive.Valid {
			project.IsActive = isActive.Bool
		} else {
			project.IsActive = true // Default is active
		}

		entry.Project = project

		// Update the appropriate map based on the operation
		switch operation {
		case string(model.OperationCreated):
			createdEntries[entry.ID] = entry
			// Remove from other maps if exists (in case of multiple operations)
			delete(updatedEntries, entry.ID)
			delete(deletedEntries, entry.ID)
		case string(model.OperationUpdated):
			// Only add to updated if not in created
			if _, exists := createdEntries[entry.ID]; !exists {
				updatedEntries[entry.ID] = entry
				// Remove from deleted if exists
				delete(deletedEntries, entry.ID)
			}
		case string(model.OperationDeleted):
			// Add to deleted if not in created or updated
			if _, existsInCreated := createdEntries[entry.ID]; !existsInCreated {
				if _, existsInUpdated := updatedEntries[entry.ID]; !existsInUpdated {
					// For soft delete, we need to ensure the entry is marked as deleted
					entry.Deleted = true
					deletedEntries[entry.ID] = entry
				}
			}
		}
	}

	// Prepare the result
	result := model.TimeEntrySyncResult{}

	// Add created entries
	for _, entry := range createdEntries {
		result.Created = append(result.Created, entry)
	}

	// Add updated entries
	for _, entry := range updatedEntries {
		result.Updated = append(result.Updated, entry)
	}

	// Add deleted entries
	for _, entry := range deletedEntries {
		result.Deleted = append(result.Deleted, entry)
	}

	return result, nil
}

// GetUpdatedProjectsOfUser retrieves projects that have been updated since a specific changelog entry ID
func (repo *postgresqlSyncRepository) GetUpdatedProjectsOfUser(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error) {
	query := `
		SELECT p.id, p.name, p.user_id, p.team_id, p.color, p.deadline::date,
			   p.hourly_rate, p.time_budget, p.is_active, p.deleted,
			   cl.operation, cl.id
		FROM change_log cl
		LEFT JOIN projects p ON cl.entity_id = p.id
		WHERE cl.id > $1
		AND cl.entity_type = 'Project'
		AND (p.user_id = $2 OR p.id IS NULL)
		AND (cl.changed_by_client != $3 OR cl.changed_by_client IS NULL OR cl.changed_by_client = '')
		ORDER BY cl.id ASC
	`

	rows, err := repo.db.Query(query, sinceTimeLogEntry, userId, excludeClientId)
	if err != nil {
		return model.ProjectSyncResult{}, err
	}
	defer rows.Close()

	// Maps to track which projects to include in each category
	createdProjects := make(map[uuid.UUID]model.Project)
	updatedProjects := make(map[uuid.UUID]model.Project)
	deletedProjects := make(map[uuid.UUID]model.Project)

	for rows.Next() {
		var project model.Project
		var operation string
		var teamID sql.NullString
		var deadline sql.NullTime
		var hourlyRate sql.NullString
		var timeBudget sql.NullInt64
		var isActive sql.NullBool
		var color sql.NullString
		var changelogID int64

		err := rows.Scan(
			&project.ID, &project.Name, &project.UserId, &teamID, &color, &deadline,
			&hourlyRate, &timeBudget, &isActive, &project.Deleted, &operation, &changelogID,
		)
		if err != nil {
			return model.ProjectSyncResult{}, err
		}

		// Set project fields
		if teamID.Valid {
			uuid, _ := uuid.FromString(teamID.String)
			project.TeamID = &uuid
		}
		if color.Valid {
			project.Color = color.String
		} else {
			project.Color = "#1E90FF" // Default blue color
		}
		if deadline.Valid {
			project.Deadline = model.NewDateOnly(deadline.Time)
		}
		if hourlyRate.Valid {
			project.HourlyRate, _ = decimal.NewFromString(hourlyRate.String)
		}
		if timeBudget.Valid {
			project.TimeBudget = int(timeBudget.Int64)
		}
		if isActive.Valid {
			project.IsActive = isActive.Bool
		} else {
			project.IsActive = true // Default is active
		}

		// Update the appropriate map based on the operation
		switch operation {
		case string(model.OperationCreated):
			createdProjects[project.ID] = project
			// Remove from other maps if exists (in case of multiple operations)
			delete(updatedProjects, project.ID)
			delete(deletedProjects, project.ID)
		case string(model.OperationUpdated):
			// Only add to updated if not in created
			if _, exists := createdProjects[project.ID]; !exists {
				updatedProjects[project.ID] = project
				// Remove from deleted if exists
				delete(deletedProjects, project.ID)
			}
		case string(model.OperationDeleted):
			// Add to deleted if not in created or updated
			if _, existsInCreated := createdProjects[project.ID]; !existsInCreated {
				if _, existsInUpdated := updatedProjects[project.ID]; !existsInUpdated {
					// For soft delete, we need to ensure the project is marked as deleted
					project.Deleted = true
					deletedProjects[project.ID] = project
				}
			}
		}
	}

	// Prepare the result
	result := model.ProjectSyncResult{}

	// Add created projects
	for _, project := range createdProjects {
		result.Created = append(result.Created, project)
	}

	// Add updated projects
	for _, project := range updatedProjects {
		result.Updated = append(result.Updated, project)
	}

	// Add deleted projects
	for _, project := range deletedProjects {
		result.Deleted = append(result.Deleted, project)
	}

	return result, nil
}

// GetProjectById retrieves a project by its ID
func (repo *postgresqlSyncRepository) GetProjectById(id uuid.UUID) (*model.Project, error) {
	query := `
		SELECT id, name, user_id, team_id, color, deadline::date, hourly_rate, time_budget, is_active
		FROM projects
		WHERE id = $1
	`

	var project model.Project
	var teamID sql.NullString
	var deadline sql.NullTime
	var hourlyRate sql.NullString
	var timeBudget sql.NullInt64
	var isActive sql.NullBool
	var color sql.NullString

	err := repo.db.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.UserId, &teamID, &color, &deadline,
		&hourlyRate, &timeBudget, &isActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Set project fields
	if teamID.Valid {
		uuid, _ := uuid.FromString(teamID.String)
		project.TeamID = &uuid
	}
	if color.Valid {
		project.Color = color.String
	} else {
		project.Color = "#1E90FF" // Default blue color
	}
	if deadline.Valid {
		project.Deadline = model.NewDateOnly(deadline.Time)
	}
	if hourlyRate.Valid {
		project.HourlyRate, _ = decimal.NewFromString(hourlyRate.String)
	}
	if timeBudget.Valid {
		project.TimeBudget = int(timeBudget.Int64)
	}
	if isActive.Valid {
		project.IsActive = isActive.Bool
	} else {
		project.IsActive = true // Default is active
	}

	return &project, nil
}

// GetTimeEntryById retrieves a time entry by its ID
func (repo *postgresqlSyncRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	query := `
		SELECT te.id, te.user_id, te.project_id, te.start_time, te.end_time, te.description,
			   p.id as project_id, p.name as project_name, p.user_id as project_user_id,
			   p.team_id as project_team_id, p.color as project_color, p.deadline::date as project_deadline,
			   p.hourly_rate as project_hourly_rate, p.time_budget as project_time_budget,
			   p.is_active as project_is_active
		FROM time_entries te
		LEFT JOIN projects p ON te.project_id = p.id
		WHERE te.id = $1
	`

	var entry model.TimeEntry
	var project model.Project
	var teamID sql.NullString
	var deadline sql.NullTime
	var hourlyRate sql.NullString
	var timeBudget sql.NullInt64
	var isActive sql.NullBool
	var color sql.NullString

	err := repo.db.QueryRow(query, id).Scan(
		&entry.ID, &entry.UserId, &entry.ProjectId, &entry.StartTime, &entry.EndTime, &entry.Description,
		&project.ID, &project.Name, &project.UserId, &teamID, &color, &deadline,
		&hourlyRate, &timeBudget, &isActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Set project fields
	if teamID.Valid {
		uuid, _ := uuid.FromString(teamID.String)
		project.TeamID = &uuid
	}
	if color.Valid {
		project.Color = color.String
	} else {
		project.Color = "#1E90FF" // Default blue color
	}
	if deadline.Valid {
		project.Deadline = model.NewDateOnly(deadline.Time)
	}
	if hourlyRate.Valid {
		project.HourlyRate, _ = decimal.NewFromString(hourlyRate.String)
	}
	if timeBudget.Valid {
		project.TimeBudget = int(timeBudget.Int64)
	}
	if isActive.Valid {
		project.IsActive = isActive.Bool
	} else {
		project.IsActive = true // Default is active
	}

	entry.Project = project

	return &entry, nil
}

// Helper methods for transaction operations

func (repo *postgresqlSyncRepository) createTimeEntry(tx *sql.Tx, entry *model.TimeEntry) error {
	query := `
		INSERT INTO time_entries (id, user_id, project_id, start_time, end_time, description, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(query, entry.ID, entry.UserId, entry.ProjectId, entry.StartTime, entry.EndTime, entry.Description, entry.Deleted)
	return err
}

func (repo *postgresqlSyncRepository) updateTimeEntry(tx *sql.Tx, entry *model.TimeEntry) error {
	query := `
		UPDATE time_entries
		SET user_id = $2, project_id = $3, start_time = $4, end_time = $5, description = $6, deleted = $7
		WHERE id = $1
	`
	_, err := tx.Exec(query, entry.ID, entry.UserId, entry.ProjectId, entry.StartTime, entry.EndTime, entry.Description, entry.Deleted)
	return err
}

func (repo *postgresqlSyncRepository) deleteTimeEntry(tx *sql.Tx, entry *model.TimeEntry) error {
	query := `
		UPDATE time_entries
		SET deleted = true
		WHERE id = $1
	`
	_, err := tx.Exec(query, entry.ID)

	if err == nil {
		entry.Deleted = true
	}
	return err
}

func (repo *postgresqlSyncRepository) createProject(tx *sql.Tx, project *model.Project) error {
	query := `
		INSERT INTO projects (id, name, user_id, team_id, color, deadline, hourly_rate, time_budget, is_active, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var deadline interface{}
	if !project.Deadline.IsZero() {
		deadline = project.Deadline.ToTime()
	} else {
		deadline = nil
	}

	_, err := tx.Exec(
		query,
		project.ID, project.Name, project.UserId, project.TeamID, project.Color,
		deadline, project.HourlyRate, project.TimeBudget, project.IsActive, project.Deleted,
	)
	return err
}

func (repo *postgresqlSyncRepository) updateProject(tx *sql.Tx, project *model.Project) error {
	query := `
		UPDATE projects
		SET name = $2, user_id = $3, team_id = $4, color = $5, deadline = $6,
		    hourly_rate = $7, time_budget = $8, is_active = $9, deleted = $10
		WHERE id = $1
	`

	var deadline interface{}
	if !project.Deadline.IsZero() {
		deadline = project.Deadline.ToTime()
	} else {
		deadline = nil
	}

	_, err := tx.Exec(
		query,
		project.ID, project.Name, project.UserId, project.TeamID, project.Color,
		deadline, project.HourlyRate, project.TimeBudget, project.IsActive, project.Deleted,
	)
	return err
}

func (repo *postgresqlSyncRepository) deleteProject(tx *sql.Tx, project *model.Project) error {
	query := `
		UPDATE projects
		SET deleted = true
		WHERE id = $1
	`
	_, err := tx.Exec(query, project.ID)

	if err == nil {
		project.Deleted = true
	}
	return err
}
