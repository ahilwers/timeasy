package postgresql

import (
	"database/sql"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
)

type postgresqlExternalConnectionRepository struct {
	db *sql.DB
}

func NewPostgreSQLExternalConnectionRepository(db *sql.DB) repository.ExternalConnectionRepository {
	return &postgresqlExternalConnectionRepository{
		db: db,
	}
}

func (repo *postgresqlExternalConnectionRepository) Create(connection *model.ExternalConnection) error {
	if connection.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		connection.ID = id
	}

	connection.CreatedAt = time.Now()
	connection.UpdatedAt = time.Now()

	query := `
		INSERT INTO external_connections (id, project_id, user_account_id, provider, project_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.db.Exec(query, connection.ID, connection.ProjectID, connection.UserAccountID, connection.Provider, connection.ProjectRef, connection.CreatedAt, connection.UpdatedAt)
	return err
}

func (repo *postgresqlExternalConnectionRepository) Update(connection *model.ExternalConnection) error {
	connection.UpdatedAt = time.Now()

	query := `
		UPDATE external_connections 
		SET user_account_id = $1, project_ref = $2, updated_at = $3
		WHERE id = $4`

	_, err := repo.db.Exec(query, connection.UserAccountID, connection.ProjectRef, connection.UpdatedAt, connection.ID)
	return err
}

func (repo *postgresqlExternalConnectionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM external_connections WHERE id = $1`
	_, err := repo.db.Exec(query, id)
	return err
}

func (repo *postgresqlExternalConnectionRepository) GetByID(id uuid.UUID) (*model.ExternalConnection, error) {
	query := `
		SELECT id, project_id, user_account_id, provider, project_ref, created_at, updated_at
		FROM external_connections 
		WHERE id = $1`

	row := repo.db.QueryRow(query, id)
	return repo.scanExternalConnection(row)
}

func (repo *postgresqlExternalConnectionRepository) GetByProjectID(projectID uuid.UUID) (*model.ExternalConnection, error) {
	query := `
		SELECT id, project_id, user_account_id, provider, project_ref, created_at, updated_at
		FROM external_connections 
		WHERE project_id = $1`

	row := repo.db.QueryRow(query, projectID)
	return repo.scanExternalConnection(row)
}

func (repo *postgresqlExternalConnectionRepository) GetByProjectIDAndProvider(projectID uuid.UUID, provider string) (*model.ExternalConnection, error) {
	query := `
		SELECT id, project_id, user_account_id, provider, project_ref, created_at, updated_at
		FROM external_connections 
		WHERE project_id = $1 AND provider = $2`

	row := repo.db.QueryRow(query, projectID, provider)
	return repo.scanExternalConnection(row)
}

func (repo *postgresqlExternalConnectionRepository) List(userID uuid.UUID) ([]*model.ExternalConnection, error) {
	query := `
		SELECT ec.id, ec.project_id, ec.user_account_id, ec.provider, ec.project_ref, ec.created_at, ec.updated_at
		FROM external_connections ec
		JOIN projects p ON ec.project_id = p.id
		WHERE p.user_id = $1
		ORDER BY ec.created_at DESC`

	rows, err := repo.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*model.ExternalConnection
	for rows.Next() {
		connection, err := repo.scanExternalConnection(rows)
		if err != nil {
			return nil, err
		}
		connections = append(connections, connection)
	}

	return connections, rows.Err()
}

func (repo *postgresqlExternalConnectionRepository) GetWithUserAccount(id uuid.UUID) (*model.ExternalConnection, error) {
	query := `
		SELECT ec.id, ec.project_id, ec.user_account_id, ec.provider, ec.project_ref, ec.created_at, ec.updated_at,
			   uea.id, uea.user_id, uea.provider, uea.account_name, uea.oauth_token, uea.base_url, uea.created_at, uea.updated_at
		FROM external_connections ec
		JOIN user_external_accounts uea ON ec.user_account_id = uea.id
		WHERE ec.id = $1`

	row := repo.db.QueryRow(query, id)
	return repo.scanExternalConnectionWithAccount(row)
}

func (repo *postgresqlExternalConnectionRepository) GetByProjectIDWithUserAccount(projectID uuid.UUID) (*model.ExternalConnection, error) {
	query := `
		SELECT ec.id, ec.project_id, ec.user_account_id, ec.provider, ec.project_ref, ec.created_at, ec.updated_at,
			   uea.id, uea.user_id, uea.provider, uea.account_name, uea.oauth_token, uea.base_url, uea.created_at, uea.updated_at
		FROM external_connections ec
		JOIN user_external_accounts uea ON ec.user_account_id = uea.id
		WHERE ec.project_id = $1`

	row := repo.db.QueryRow(query, projectID)
	return repo.scanExternalConnectionWithAccount(row)
}

func (repo *postgresqlExternalConnectionRepository) scanExternalConnection(scanner scanner) (*model.ExternalConnection, error) {
	connection := &model.ExternalConnection{}
	err := scanner.Scan(
		&connection.ID,
		&connection.ProjectID,
		&connection.UserAccountID,
		&connection.Provider,
		&connection.ProjectRef,
		&connection.CreatedAt,
		&connection.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (repo *postgresqlExternalConnectionRepository) scanExternalConnectionWithAccount(scanner scanner) (*model.ExternalConnection, error) {
	connection := &model.ExternalConnection{}
	account := &model.UserExternalAccount{}
	var baseURL sql.NullString
	
	err := scanner.Scan(
		&connection.ID,
		&connection.ProjectID,
		&connection.UserAccountID,
		&connection.Provider,
		&connection.ProjectRef,
		&connection.CreatedAt,
		&connection.UpdatedAt,
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.AccountName,
		&account.OAuthToken,
		&baseURL,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	if baseURL.Valid {
		account.BaseURL = baseURL.String
	}
	
	connection.UserAccount = account
	return connection, nil
}