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

func (repo *postgresqlTimeEntryRepository) AddTimeEntry(entry *model.TimeEntry) error {
	if entry.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		entry.ID = id
	}

	query := `
		INSERT INTO time_entries (
			id, user_id, project_id, start_time, end_time, description
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := repo.db.Exec(
		query,
		entry.ID,
		entry.UserId,
		entry.ProjectId,
		entry.StartTime,
		entry.EndTime,
		entry.Description,
	)

	return err
}

func (repo *postgresqlTimeEntryRepository) AddTimeEntryList(entries []model.TimeEntry) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO time_entries (
			id, user_id, project_id, start_time, end_time, description
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, entry := range entries {
		if entry.ID == uuid.Nil {
			id, err := uuid.NewV4()
			if err != nil {
				tx.Rollback()
				return err
			}
			entry.ID = id
		}

		_, err := tx.Exec(
			query,
			entry.ID,
			entry.UserId,
			entry.ProjectId,
			entry.StartTime,
			entry.EndTime,
			entry.Description,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (repo *postgresqlTimeEntryRepository) UpdateTimeEntry(entry *model.TimeEntry) error {
	query := `
		UPDATE time_entries SET
			user_id = $1,
			project_id = $2,
			start_time = $3,
			end_time = $4,
			description = $5
		WHERE id = $6
	`

	result, err := repo.db.Exec(
		query,
		entry.UserId,
		entry.ProjectId,
		entry.StartTime,
		entry.EndTime,
		entry.Description,
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

func (repo *postgresqlTimeEntryRepository) UpdateTimeEntryList(entries []model.TimeEntry) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	query := `
		UPDATE time_entries SET
			user_id = $1,
			project_id = $2,
			start_time = $3,
			end_time = $4,
			description = $5
		WHERE id = $6
	`

	for _, entry := range entries {
		result, err := tx.Exec(
			query,
			entry.UserId,
			entry.ProjectId,
			entry.StartTime,
			entry.EndTime,
			entry.Description,
			entry.ID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		if rowsAffected == 0 {
			tx.Rollback()
			return repository.ErrEntityNotFound
		}
	}

	return tx.Commit()
}

func (repo *postgresqlTimeEntryRepository) DeleteTimeEntry(entry *model.TimeEntry) error {
	query := `DELETE FROM time_entries WHERE id = $1`
	result, err := repo.db.Exec(query, entry.ID)
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

func (repo *postgresqlTimeEntryRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description
		FROM time_entries
		WHERE id = $1
	`

	var entry model.TimeEntry
	err := repo.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.UserId,
		&entry.ProjectId,
		&entry.StartTime,
		&entry.EndTime,
		&entry.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	return &entry, nil
}

func (repo *postgresqlTimeEntryRepository) GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description
		FROM time_entries
		WHERE user_id = $1 AND (end_time IS NULL OR end_time = '0001-01-01 00:00:00'::timestamp)
		ORDER BY start_time DESC
		LIMIT 1
	`

	var entry model.TimeEntry
	err := repo.db.QueryRow(query, userId).Scan(
		&entry.ID,
		&entry.UserId,
		&entry.ProjectId,
		&entry.StartTime,
		&entry.EndTime,
		&entry.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEntityNotFound
		}
		return nil, err
	}

	return &entry, nil
}

func (repo *postgresqlTimeEntryRepository) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description
		FROM time_entries
		WHERE user_id = $1
		ORDER BY start_time DESC, end_time DESC
	`

	return repo.queryTimeEntries(query, userId)
}

func (repo *postgresqlTimeEntryRepository) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	query := `
		SELECT
			id, user_id, project_id, start_time, end_time, description
		FROM time_entries
		WHERE user_id = $1 AND project_id = $2
		ORDER BY start_time DESC, end_time DESC
	`

	return repo.queryTimeEntries(query, userId, projectId)
}

func (repo *postgresqlTimeEntryRepository) GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error) {
	query := `
        SELECT
            id, user_id, project_id, start_time, end_time, description
        FROM time_entries
        WHERE user_id = $1
          AND DATE(start_time) >= $2
          AND (end_time IS NULL OR DATE(end_time) <= $3)
    `

	args := []interface{}{userId, startDate, endDate}

	if projectId != uuid.Nil {
		query += ` AND project_id = $4`
		args = append(args, projectId)
	}

	query += ` ORDER BY start_time DESC, end_time DESC`

	return repo.queryTimeEntries(query, args...)
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
		err := rows.Scan(
			&entry.ID,
			&entry.UserId,
			&entry.ProjectId,
			&entry.StartTime,
			&entry.EndTime,
			&entry.Description,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
