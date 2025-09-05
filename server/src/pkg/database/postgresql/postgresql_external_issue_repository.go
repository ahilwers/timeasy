package postgresql

import (
	"database/sql"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
)

type postgresqlExternalIssueRepository struct {
	db *sql.DB
}

func NewPostgreSQLExternalIssueRepository(db *sql.DB) repository.ExternalIssueRepository {
	return &postgresqlExternalIssueRepository{
		db: db,
	}
}

func (repo *postgresqlExternalIssueRepository) Create(issue *model.ExternalIssue) error {
	if issue.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		issue.ID = id
	}

	issue.CreatedAt = time.Now()
	issue.UpdatedAt = time.Now()

	query := `
		INSERT INTO external_issues (id, project_id, provider, key_or_number, title, state, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (project_id, provider, key_or_number) 
		DO UPDATE SET 
			title = EXCLUDED.title,
			state = EXCLUDED.state,
			url = EXCLUDED.url,
			updated_at = EXCLUDED.updated_at`

	_, err := repo.db.Exec(query, issue.ID, issue.ProjectID, issue.Provider, issue.KeyOrNumber, issue.Title, issue.State, issue.URL, issue.CreatedAt, issue.UpdatedAt)
	return err
}

func (repo *postgresqlExternalIssueRepository) Update(issue *model.ExternalIssue) error {
	issue.UpdatedAt = time.Now()

	query := `
		UPDATE external_issues 
		SET title = $1, state = $2, url = $3, updated_at = $4
		WHERE id = $5`

	_, err := repo.db.Exec(query, issue.Title, issue.State, issue.URL, issue.UpdatedAt, issue.ID)
	return err
}

func (repo *postgresqlExternalIssueRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM external_issues WHERE id = $1`
	_, err := repo.db.Exec(query, id)
	return err
}

func (repo *postgresqlExternalIssueRepository) GetByID(id uuid.UUID) (*model.ExternalIssue, error) {
	query := `
		SELECT id, project_id, provider, key_or_number, title, state, url, created_at, updated_at
		FROM external_issues 
		WHERE id = $1`

	row := repo.db.QueryRow(query, id)
	return repo.scanExternalIssue(row)
}

func (repo *postgresqlExternalIssueRepository) GetByProjectIDAndKey(projectID uuid.UUID, provider string, keyOrNumber string) (*model.ExternalIssue, error) {
	query := `
		SELECT id, project_id, provider, key_or_number, title, state, url, created_at, updated_at
		FROM external_issues 
		WHERE project_id = $1 AND provider = $2 AND key_or_number = $3`

	row := repo.db.QueryRow(query, projectID, provider, keyOrNumber)
	return repo.scanExternalIssue(row)
}

func (repo *postgresqlExternalIssueRepository) List(projectID uuid.UUID) ([]*model.ExternalIssue, error) {
	query := `
		SELECT id, project_id, provider, key_or_number, title, state, url, created_at, updated_at
		FROM external_issues 
		WHERE project_id = $1
		ORDER BY updated_at DESC`

	rows, err := repo.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*model.ExternalIssue
	for rows.Next() {
		issue, err := repo.scanExternalIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

func (repo *postgresqlExternalIssueRepository) ListByProvider(projectID uuid.UUID, provider string) ([]*model.ExternalIssue, error) {
	query := `
		SELECT id, project_id, provider, key_or_number, title, state, url, created_at, updated_at
		FROM external_issues 
		WHERE project_id = $1 AND provider = $2
		ORDER BY updated_at DESC`

	rows, err := repo.db.Query(query, projectID, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []*model.ExternalIssue
	for rows.Next() {
		issue, err := repo.scanExternalIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

func (repo *postgresqlExternalIssueRepository) BatchUpsert(issues []*model.ExternalIssue) error {
	if len(issues) == 0 {
		return nil
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO external_issues (id, project_id, provider, key_or_number, title, state, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (project_id, provider, key_or_number) 
		DO UPDATE SET 
			title = EXCLUDED.title,
			state = EXCLUDED.state,
			url = EXCLUDED.url,
			updated_at = EXCLUDED.updated_at`

	for _, issue := range issues {
		if issue.ID == uuid.Nil {
			id, err := uuid.NewV4()
			if err != nil {
				return err
			}
			issue.ID = id
		}

		if issue.CreatedAt.IsZero() {
			issue.CreatedAt = time.Now()
		}
		issue.UpdatedAt = time.Now()

		_, err := tx.Exec(query, issue.ID, issue.ProjectID, issue.Provider, issue.KeyOrNumber, issue.Title, issue.State, issue.URL, issue.CreatedAt, issue.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *postgresqlExternalIssueRepository) DeleteOldIssues(projectID uuid.UUID, provider string, cutoffTime int64) error {
	query := `
		DELETE FROM external_issues 
		WHERE project_id = $1 AND provider = $2 AND updated_at < to_timestamp($3)`

	_, err := repo.db.Exec(query, projectID, provider, cutoffTime)
	return err
}

func (repo *postgresqlExternalIssueRepository) scanExternalIssue(scanner scanner) (*model.ExternalIssue, error) {
	issue := &model.ExternalIssue{}
	err := scanner.Scan(
		&issue.ID,
		&issue.ProjectID,
		&issue.Provider,
		&issue.KeyOrNumber,
		&issue.Title,
		&issue.State,
		&issue.URL,
		&issue.CreatedAt,
		&issue.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return issue, nil
}