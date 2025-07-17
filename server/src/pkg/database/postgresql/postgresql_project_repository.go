package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
)

type postgresqlProjectRepository struct {
	db             *sql.DB
	teamRepository repository.TeamRepository
}

func NewPostgreSQLProjectRepository(db *sql.DB, teamRepository repository.TeamRepository) repository.ProjectRepository {
	return &postgresqlProjectRepository{
		db:             db,
		teamRepository: teamRepository,
	}
}

func (repo *postgresqlProjectRepository) BeginTransaction() (model.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (repo *postgresqlProjectRepository) AddProject(project *model.Project, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	if project.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		project.ID = id
	}

	query := `
		INSERT INTO projects (
			id, name, user_id, team_id, color, hourly_rate, time_budget, deadline, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var teamID *uuid.UUID
	if project.TeamID != nil && *project.TeamID != uuid.Nil {
		teamID = project.TeamID
	}

	// Set default values if not provided
	if project.Color == "" {
		project.Color = "#1E90FF"
	}

	var deadline interface{}
	if !project.Deadline.IsZero() {
		deadline = project.Deadline.ToTime()
	} else {
		deadline = nil
	}

	_, err := sqlTx.Exec(query,
		project.ID,
		project.Name,
		project.UserId,
		teamID,
		project.Color,
		project.HourlyRate,
		project.TimeBudget,
		deadline,
		project.IsActive,
	)

	return err
}

func (repo *postgresqlProjectRepository) GetProjectById(id uuid.UUID) (*model.Project, error) {
	query := `
		SELECT
			id, name, user_id, team_id, color, hourly_rate, time_budget, deadline::date, is_active
		FROM projects
		WHERE id = $1
	`

	var project model.Project
	var teamID uuid.NullUUID

	err := repo.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.UserId,
		&teamID,
		&project.Color,
		&project.HourlyRate,
		&project.TimeBudget,
		&project.Deadline,
		&project.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	if teamID.Valid {
		project.TeamID = &teamID.UUID
	}

	return &project, nil
}

func (repo *postgresqlProjectRepository) UpdateProject(project *model.Project, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `
		UPDATE projects SET
			name = $1,
			user_id = $2,
			team_id = $3,
			color = $4,
			hourly_rate = $5,
			time_budget = $6,
			deadline = $7,
			is_active = $8
		WHERE id = $9
	`

	var teamID *uuid.UUID
	if project.TeamID != nil && *project.TeamID != uuid.Nil {
		teamID = project.TeamID
	}

	var deadline interface{}
	if !project.Deadline.IsZero() {
		deadline = project.Deadline.ToTime()
	} else {
		deadline = nil
	}

	result, err := sqlTx.Exec(
		query,
		project.Name,
		project.UserId,
		teamID,
		project.Color,
		project.HourlyRate,
		project.TimeBudget,
		deadline,
		project.IsActive,
		project.ID,
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

func (repo *postgresqlProjectRepository) DeleteProject(project *model.Project, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `DELETE FROM projects WHERE id = $1`
	result, err := sqlTx.Exec(query, project.ID)
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

func (repo *postgresqlProjectRepository) GetAllProjects() ([]model.Project, error) {
	query := `
		SELECT
			id, name, user_id, team_id, color, hourly_rate, time_budget, deadline::date, is_active
		FROM projects
		ORDER BY name
	`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (repo *postgresqlProjectRepository) GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error) {
	var err error
	teamIds, err := repo.getTeamIdsOfUser(userId)
	if err != nil {
		return nil, err
	}

	var query string
	var args []interface{}
	var rows *sql.Rows

	if len(teamIds) > 0 {
		args := make([]interface{}, 0, len(teamIds)+1)
		args = append(args, userId)

		placeholders := make([]string, len(teamIds))
		for i, teamID := range teamIds {
			args = append(args, teamID)
			placeholders[i] = fmt.Sprintf("$%d", i+2) // $2, $3, etc.
		}

		query = fmt.Sprintf(`
            SELECT
                id, name, user_id, team_id, color, hourly_rate, time_budget, deadline, is_active
            FROM projects
            WHERE user_id = $1 OR team_id = ANY(ARRAY[%s]::uuid[])
            ORDER BY name
        `, strings.Join(placeholders, ","))

		rows, err = repo.db.Query(query, args...)
	} else {
		query = `
			SELECT
				id, name, user_id, team_id, color, hourly_rate, time_budget, deadline::date, is_active
			FROM projects
			WHERE user_id = $1
			ORDER BY name
		`
		args = []interface{}{userId}
		rows, err = repo.db.Query(query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (repo *postgresqlProjectRepository) getTeamIdsOfUser(userId uuid.UUID) ([]uuid.UUID, error) {
	teamAssignments, err := repo.teamRepository.GetTeamsOfUser(userId)
	if err != nil {
		return nil, err
	}

	teamIds := make([]uuid.UUID, 0, len(teamAssignments))
	for _, assignment := range teamAssignments {
		teamIds = append(teamIds, assignment.TeamID)
	}

	return teamIds, nil
}

func scanProject(rows *sql.Rows) (*model.Project, error) {
	var project model.Project
	var teamID uuid.NullUUID
	var color sql.NullString
	var hourlyRate sql.NullFloat64

	err := rows.Scan(
		&project.ID,
		&project.Name,
		&project.UserId,
		&teamID,
		&color,
		&hourlyRate,
		&project.TimeBudget,
		&project.Deadline,
		&project.IsActive,
	)

	if err != nil {
		return nil, err
	}

	if teamID.Valid {
		project.TeamID = &teamID.UUID
	}

	if color.Valid {
		project.Color = color.String
	}

	if hourlyRate.Valid {
		project.HourlyRate = decimal.NewFromFloat(hourlyRate.Float64)
	}

	return &project, nil
}
