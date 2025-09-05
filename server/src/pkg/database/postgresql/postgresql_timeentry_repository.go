package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type postgresqlTimeEntryRepository struct {
	db *sql.DB
}

func NewPostgreSQLTimeEntryRepository(db *sql.DB) repository.TimeEntryRepository {
	return &postgresqlTimeEntryRepository{
		db: db,
	}
}

func (repo *postgresqlTimeEntryRepository) BeginTransaction() (model.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (repo *postgresqlTimeEntryRepository) AddTimeEntry(entry *model.TimeEntry, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	if entry.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		entry.ID = id
	}

	query := `
		INSERT INTO time_entries (
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := sqlTx.Exec(
		query,
		entry.ID,
		entry.UserId,
		entry.ProjectId,
		entry.StartTime,
		entry.EndTime,
		entry.Description,
		entry.ExternalIssueID,
		entry.PendingExternalRef,
	)

	return err
}

func (repo *postgresqlTimeEntryRepository) AddTimeEntryList(entries []model.TimeEntry, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}

	query := `
		INSERT INTO time_entries (
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, entry := range entries {
		if entry.ID == uuid.Nil {
			id, err := uuid.NewV4()
			if err != nil {
				sqlTx.Rollback()
				return err
			}
			entry.ID = id
		}

		_, err := sqlTx.Exec(
			query,
			entry.ID,
			entry.UserId,
			entry.ProjectId,
			entry.StartTime,
			entry.EndTime,
			entry.Description,
			entry.ExternalIssueID,
			entry.PendingExternalRef,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *postgresqlTimeEntryRepository) UpdateTimeEntry(entry *model.TimeEntry, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `
		UPDATE time_entries SET
			user_id = $1,
			project_id = $2,
			start_time = $3,
			end_time = $4,
			description = $5,
			external_issue_id = $6,
			pending_external_ref = $7
		WHERE id = $8
	`

	result, err := sqlTx.Exec(
		query,
		entry.UserId,
		entry.ProjectId,
		entry.StartTime,
		entry.EndTime,
		entry.Description,
		entry.ExternalIssueID,
		entry.PendingExternalRef,
		entry.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrEntityNotFound
	}

	return nil
}

func (repo *postgresqlTimeEntryRepository) UpdateTimeEntryList(entries []model.TimeEntry, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}

	query := `
		UPDATE time_entries SET
			user_id = $1,
			project_id = $2,
			start_time = $3,
			end_time = $4,
			description = $5,
			external_issue_id = $6,
			pending_external_ref = $7
		WHERE id = $8
	`

	for _, entry := range entries {
		result, err := sqlTx.Exec(
			query,
			entry.UserId,
			entry.ProjectId,
			entry.StartTime,
			entry.EndTime,
			entry.Description,
			entry.ExternalIssueID,
			entry.PendingExternalRef,
			entry.ID,
		)

		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return repository.ErrEntityNotFound
		}
	}

	return nil
}

func (repo *postgresqlTimeEntryRepository) DeleteTimeEntry(entry *model.TimeEntry, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `UPDATE time_entries SET deleted = true WHERE id = $1`
	result, err := sqlTx.Exec(query, entry.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrEntityNotFound
	}

	entry.Deleted = true
	return nil
}

func (repo *postgresqlTimeEntryRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE id = $1 AND deleted = false
	`

	var entry model.TimeEntry
	var pendingExternalRef sql.NullString
	err := repo.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.UserId,
		&entry.ProjectId,
		&entry.StartTime,
		&entry.EndTime,
		&entry.Description,
		&entry.ExternalIssueID,
		&pendingExternalRef,
	)
	
	// Handle nullable string
	if pendingExternalRef.Valid {
		entry.PendingExternalRef = &pendingExternalRef.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	// Ensure times have UTC location
	repo.setupTimeLocation(&entry)

	return &entry, nil
}

func (repo *postgresqlTimeEntryRepository) GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE user_id = $1 AND (end_time IS NULL OR end_time = '0001-01-01 00:00:00'::timestamp) AND deleted = false
		ORDER BY start_time DESC
		LIMIT 1
	`

	var entry model.TimeEntry
	var pendingExternalRef sql.NullString
	err := repo.db.QueryRow(query, userId).Scan(
		&entry.ID,
		&entry.UserId,
		&entry.ProjectId,
		&entry.StartTime,
		&entry.EndTime,
		&entry.Description,
		&entry.ExternalIssueID,
		&pendingExternalRef,
	)
	
	// Handle nullable string
	if pendingExternalRef.Valid {
		entry.PendingExternalRef = &pendingExternalRef.String
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	// Ensure times have UTC location
	repo.setupTimeLocation(&entry)

	return &entry, nil
}

func (repo *postgresqlTimeEntryRepository) GetAllTimeEntries() ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE deleted = false
		ORDER BY start_time DESC, end_time DESC
	`

	return repo.queryTimeEntries(query)
}

func (repo *postgresqlTimeEntryRepository) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE user_id = $1 AND deleted = false
		ORDER BY start_time DESC, end_time DESC
	`

	return repo.queryTimeEntries(query, userId)
}

func (repo *postgresqlTimeEntryRepository) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE user_id = $1 AND project_id = $2 AND deleted = false
		ORDER BY start_time DESC, end_time DESC
	`

	return repo.queryTimeEntries(query, userId, projectId)
}

func (repo *postgresqlTimeEntryRepository) GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error) {
	query := `
        SELECT
            id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
        FROM time_entries
        WHERE user_id = $1
          AND DATE(start_time) >= $2
          AND (end_time IS NULL OR DATE(end_time) <= $3)
          AND deleted = false
    `

	args := []interface{}{userId, startDate, endDate}

	if projectId != uuid.Nil {
		query += ` AND project_id = $4`
		args = append(args, projectId)
	}

	query += ` ORDER BY start_time DESC, end_time DESC`

	return repo.queryTimeEntries(query, args...)
}

func (repo *postgresqlTimeEntryRepository) GetOpenTimeEntriesForProject(userId uuid.UUID, projectId uuid.UUID, tx model.Transaction) ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description, external_issue_id, pending_external_ref
		FROM time_entries
		WHERE user_id = $1 AND project_id = $2 AND (end_time IS NULL OR end_time = '0001-01-01 00:00:00'::timestamp) AND deleted = false
		ORDER BY start_time ASC
	`

	var rows *sql.Rows
	var err error
	
	if tx != nil {
		sqlTx, ok := tx.(*sql.Tx)
		if !ok {
			return nil, errors.New("invalid transaction type")
		}
		rows, err = sqlTx.Query(query, userId, projectId)
	} else {
		rows, err = repo.db.Query(query, userId, projectId)
	}
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.TimeEntry
	for rows.Next() {
		var entry model.TimeEntry
		var pendingExternalRef sql.NullString
		err := rows.Scan(
			&entry.ID,
			&entry.UserId,
			&entry.ProjectId,
			&entry.StartTime,
			&entry.EndTime,
			&entry.Description,
			&entry.ExternalIssueID,
			&pendingExternalRef,
		)
		if err != nil {
			return nil, err
		}
		
		// Handle nullable string
		if pendingExternalRef.Valid {
			entry.PendingExternalRef = &pendingExternalRef.String
		}

		// Ensure times have UTC location
		repo.setupTimeLocation(&entry)

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// setupTimeLocation ensures that all time fields in the TimeEntry have their location set to UTC
func (repo *postgresqlTimeEntryRepository) setupTimeLocation(entry *model.TimeEntry) {
	if !entry.StartTime.IsZero() {
		entry.StartTime = entry.StartTime.In(time.UTC)
	}
	if !entry.EndTime.IsZero() {
		entry.EndTime = entry.EndTime.In(time.UTC)
	}
}

func (repo *postgresqlTimeEntryRepository) queryTimeEntries(query string, args ...interface{}) ([]model.TimeEntry, error) {
	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.TimeEntry
	for rows.Next() {
		var entry model.TimeEntry
		var pendingExternalRef sql.NullString
		err := rows.Scan(
			&entry.ID,
			&entry.UserId,
			&entry.ProjectId,
			&entry.StartTime,
			&entry.EndTime,
			&entry.Description,
			&entry.ExternalIssueID,
			&pendingExternalRef,
		)
		if err != nil {
			return nil, err
		}
		
		// Handle nullable string
		if pendingExternalRef.Valid {
			entry.PendingExternalRef = &pendingExternalRef.String
		}

		// Ensure times have UTC location
		repo.setupTimeLocation(&entry)

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// GetLastActivityTimeForProject returns the most recent activity time for a project
func (repo *postgresqlTimeEntryRepository) GetLastActivityTimeForProject(projectId uuid.UUID) (time.Time, error) {
	query := `
		SELECT MAX(GREATEST(
			COALESCE(start_time, '1970-01-01'::timestamp),
			COALESCE(end_time, '1970-01-01'::timestamp)
		)) as last_activity
		FROM time_entries
		WHERE project_id = $1 AND deleted = false
	`
	
	var lastActivity sql.NullTime
	err := repo.db.QueryRow(query, projectId).Scan(&lastActivity)
	if err != nil {
		return time.Time{}, err
	}
	
	if !lastActivity.Valid {
		return time.Time{}, nil // No activity found
	}
	
	return lastActivity.Time.In(time.UTC), nil
}

// GetProjectsWithRecentActivity returns project IDs that have had activity since the given time
func (repo *postgresqlTimeEntryRepository) GetProjectsWithRecentActivity(since time.Time) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT project_id
		FROM time_entries
		WHERE (start_time >= $1 OR end_time >= $1) AND deleted = false
		ORDER BY project_id
	`
	
	rows, err := repo.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projectIDs []uuid.UUID
	for rows.Next() {
		var projectID uuid.UUID
		err := rows.Scan(&projectID)
		if err != nil {
			return nil, err
		}
		projectIDs = append(projectIDs, projectID)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return projectIDs, nil
}
