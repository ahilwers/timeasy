package postgresql

import (
	"database/sql"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	_ "github.com/lib/pq"
)

type postgresqlUserExternalAccountRepository struct {
	db *sql.DB
}

func NewPostgreSQLUserExternalAccountRepository(db *sql.DB) repository.UserExternalAccountRepository {
	return &postgresqlUserExternalAccountRepository{
		db: db,
	}
}

func (repo *postgresqlUserExternalAccountRepository) Create(account *model.UserExternalAccount) error {
	if account.ID == uuid.Nil {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		account.ID = id
	}

	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_external_accounts (id, user_id, provider, account_name, oauth_token, base_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := repo.db.Exec(query, account.ID, account.UserID, account.Provider, account.AccountName, account.OAuthToken, account.BaseURL, account.CreatedAt, account.UpdatedAt)
	return err
}

func (repo *postgresqlUserExternalAccountRepository) Update(account *model.UserExternalAccount) error {
	account.UpdatedAt = time.Now()

	query := `
		UPDATE user_external_accounts 
		SET account_name = $1, oauth_token = $2, base_url = $3, updated_at = $4
		WHERE id = $5`

	_, err := repo.db.Exec(query, account.AccountName, account.OAuthToken, account.BaseURL, account.UpdatedAt, account.ID)
	return err
}

func (repo *postgresqlUserExternalAccountRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM user_external_accounts WHERE id = $1`
	_, err := repo.db.Exec(query, id)
	return err
}

func (repo *postgresqlUserExternalAccountRepository) GetByID(id uuid.UUID) (*model.UserExternalAccount, error) {
	query := `
		SELECT id, user_id, provider, account_name, oauth_token, base_url, created_at, updated_at
		FROM user_external_accounts 
		WHERE id = $1`

	row := repo.db.QueryRow(query, id)
	return repo.scanUserExternalAccount(row)
}

func (repo *postgresqlUserExternalAccountRepository) GetByUserIDAndProvider(userID uuid.UUID, provider string) ([]*model.UserExternalAccount, error) {
	query := `
		SELECT id, user_id, provider, account_name, oauth_token, base_url, created_at, updated_at
		FROM user_external_accounts 
		WHERE user_id = $1 AND provider = $2
		ORDER BY account_name`

	rows, err := repo.db.Query(query, userID, provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.UserExternalAccount
	for rows.Next() {
		account, err := repo.scanUserExternalAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (repo *postgresqlUserExternalAccountRepository) GetByUserID(userID uuid.UUID) ([]*model.UserExternalAccount, error) {
	query := `
		SELECT id, user_id, provider, account_name, oauth_token, base_url, created_at, updated_at
		FROM user_external_accounts 
		WHERE user_id = $1
		ORDER BY provider, account_name`

	rows, err := repo.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.UserExternalAccount
	for rows.Next() {
		account, err := repo.scanUserExternalAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (repo *postgresqlUserExternalAccountRepository) ValidateAccountOwnership(accountID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM user_external_accounts WHERE id = $1 AND user_id = $2`
	
	var count int
	err := repo.db.QueryRow(query, accountID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (repo *postgresqlUserExternalAccountRepository) scanUserExternalAccount(scanner scanner) (*model.UserExternalAccount, error) {
	account := &model.UserExternalAccount{}
	var baseURL sql.NullString
	
	err := scanner.Scan(
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
	
	return account, nil
}