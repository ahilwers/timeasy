package postgresql

import (
	"database/sql"
	"errors"
	"strings"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type postgresqlTeamRepository struct {
	db *sql.DB
}

func NewPostgreSQLTeamRepository(db *sql.DB) repository.TeamRepository {
	return &postgresqlTeamRepository{
		db: db,
	}
}

func (repo *postgresqlTeamRepository) BeginTransaction() (model.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (repo *postgresqlTeamRepository) AddTeam(team *model.Team) error {
	if team.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		team.ID = id
	}

	query := `
		INSERT INTO teams (id, name1, name2, name3)
		VALUES ($1, $2, $3, $4)
	`

	_, err := repo.db.Exec(
		query,
		team.ID,
		team.Name1,
		team.Name2,
		team.Name3,
	)
	return err
}

func (repo *postgresqlTeamRepository) UpdateTeam(team *model.Team) error {
	query := `
		UPDATE teams
		SET name1 = $1, name2 = $2, name3 = $3
		WHERE id = $4
	`

	result, err := repo.db.Exec(
		query,
		team.Name1,
		team.Name2,
		team.Name3,
		team.ID,
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

func (repo *postgresqlTeamRepository) DeleteTeam(team *model.Team, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `DELETE FROM teams WHERE id = $1`

	result, err := sqlTx.Exec(query, team.ID)
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

func (repo *postgresqlTeamRepository) GetTeamById(id uuid.UUID) (*model.Team, error) {
	query := `
		SELECT id, name1, name2, name3
		FROM teams
		WHERE id = $1
	`

	var team model.Team
	err := repo.db.QueryRow(query, id).Scan(
		&team.ID,
		&team.Name1,
		&team.Name2,
		&team.Name3,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	return &team, nil
}

func (repo *postgresqlTeamRepository) GetAllTeams() ([]model.Team, error) {
	query := `
		SELECT id, name1, name2, name3
		FROM teams
		ORDER BY name1, name2, name3
	`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []model.Team
	for rows.Next() {
		var team model.Team
		err := rows.Scan(
			&team.ID,
			&team.Name1,
			&team.Name2,
			&team.Name3,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

func (repo *postgresqlTeamRepository) AddUserTeamAssignment(assignment *model.UserTeamAssignment) error {
	if assignment.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		assignment.ID = id
	}

	query := `
		INSERT INTO user_team_assignments (id, user_id, team_id, roles)
		VALUES ($1, $2, $3, $4)
	`

	roles := strings.Join(assignment.Roles, ",")

	_, err := repo.db.Exec(
		query,
		assignment.ID,
		assignment.UserID.String(),
		assignment.TeamID,
		roles,
	)

	return err
}

func (repo *postgresqlTeamRepository) GetTeamsOfUser(userID uuid.UUID) ([]model.UserTeamAssignment, error) {
	query := `
		SELECT uta.id, uta.user_id, uta.team_id, uta.roles,
		       t.name1, t.name2, t.name3
		FROM user_team_assignments uta
		JOIN teams t ON uta.team_id = t.id
		WHERE uta.user_id = $1
		ORDER BY t.name1, t.name2, t.name3
	`

	rows, err := repo.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []model.UserTeamAssignment
	for rows.Next() {
		var assignment model.UserTeamAssignment
		var rolesStr string

		err := rows.Scan(
			&assignment.ID,
			&assignment.UserID,
			&assignment.TeamID,
			&rolesStr,
			&assignment.Team.Name1,
			&assignment.Team.Name2,
			&assignment.Team.Name3,
		)
		if err != nil {
			return nil, err
		}

		assignment.Roles = strings.Split(rolesStr, ",")
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (repo *postgresqlTeamRepository) GetUserTeamAssignment(userID, teamID uuid.UUID) (*model.UserTeamAssignment, error) {
	query := `
		SELECT uta.id, uta.user_id, uta.team_id, uta.roles,
		       t.name1, t.name2, t.name3
		FROM user_team_assignments uta
		JOIN teams t ON uta.team_id = t.id
		WHERE uta.user_id = $1 AND uta.team_id = $2
	`

	var assignment model.UserTeamAssignment
	var rolesStr string

	err := repo.db.QueryRow(query, userID, teamID).Scan(
		&assignment.ID,
		&assignment.UserID,
		&assignment.TeamID,
		&rolesStr,
		&assignment.Team.Name1,
		&assignment.Team.Name2,
		&assignment.Team.Name3,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	assignment.Roles = strings.Split(rolesStr, ",")
	return &assignment, nil
}

func (repo *postgresqlTeamRepository) DeleteUserTeamAssignment(assignment *model.UserTeamAssignment) error {
	query := `
		DELETE FROM user_team_assignments
		WHERE user_id = $1 AND team_id = $2
	`

	result, err := repo.db.Exec(query, assignment.UserID, assignment.TeamID)
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
func (repo *postgresqlTeamRepository) DeleteAllUserAssignmentsOfTeam(teamId uuid.UUID, tx model.Transaction) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}
	query := `
		DELETE FROM user_team_assignments
		WHERE team_id = $1
	`

	_, err := sqlTx.Exec(query, teamId)
	if err != nil {
		return err
	}
	return nil
}

func (repo *postgresqlTeamRepository) UpdateUserTeamAssignment(assignment *model.UserTeamAssignment) error {
	query := `
		UPDATE user_team_assignments
		SET roles = $1
		WHERE user_id = $2 AND team_id = $3
	`

	roles := strings.Join(assignment.Roles, ",")

	_, err := repo.db.Exec(
		query,
		roles,
		assignment.UserID,
		assignment.TeamID,
	)
	return err

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrEntityNotFound
		}
		return err
	}

	return nil
}
