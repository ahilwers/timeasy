package postgresql

import (
    "database/sql"
    "errors"
    "fmt"
    "github.com/gofrs/uuid"
    "time"

    "timeasy-server/pkg/domain/model"
    "timeasy-server/pkg/domain/repository"
)

type postgresqlChangelogRepository struct {
    db *sql.DB
}

func NewPostgreSQLChangelogRepository(db *sql.DB) repository.ChangelogRepository {
    return &postgresqlChangelogRepository{
        db: db,
    }
}

func (r *postgresqlChangelogRepository) AddChangelogEntry(entry *model.ChangelogEntry, tx model.Transaction) error {
    sqlTx, ok := tx.(*sql.Tx)
    if !ok {
        return errors.New("invalid transaction type")
    }
    query := `
		INSERT INTO change_log (
			entity_type,
			entity_id,
			operation,
			changed_by_user,
			changed_by_client,
			changed_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

    if entry.ChangedAt.IsZero() {
        entry.ChangedAt = time.Now().UTC()
    }

    _, err := sqlTx.Exec(
        query,
        entry.EntityType,
        entry.EntityID,
        entry.Operation,
        entry.ChangedByUser,
        entry.ChangedByClient, //kann nil sein
        entry.ChangedAt,
    )

    return err
}

func (r *postgresqlChangelogRepository) GetChangelogEntries(filter *repository.ChangelogFilter) ([]*model.ChangelogEntry, error) {
    query := `
		SELECT
			id, entity_type, entity_id, operation,
			changed_by_user, changed_by_client, changed_at
		FROM change_log ORDER BY id
	`

    var args []interface{}
    argPos := 1

    if filter != nil {
        if filter.EntityType != "" {
            query += fmt.Sprintf(" AND entity_type = $%d", argPos)
            args = append(args, filter.EntityType)
            argPos++
        }

        if filter.EntityID != uuid.Nil {
            query += fmt.Sprintf(" AND entity_id = $%d", argPos)
            args = append(args, filter.EntityID)
            argPos++
        }

        if !filter.StartDate.IsZero() {
            query += fmt.Sprintf(" AND changed_at >= $%d", argPos)
            args = append(args, filter.StartDate)
            argPos++
        }

        if !filter.EndDate.IsZero() {
            query += fmt.Sprintf(" AND changed_at <= $%d", argPos)
            args = append(args, filter.EndDate)
            argPos++
        }

        if filter.ChangedByUser != uuid.Nil {
            query += fmt.Sprintf(" AND changed_by_user = $%d", argPos)
            args = append(args, filter.ChangedByUser)
            argPos++
        }

        if filter.Operation != "" {
            query += fmt.Sprintf(" AND change_type = $%d", argPos)
            args = append(args, filter.Operation)
            argPos++
        }

        query += " ORDER BY changed_at DESC"

        if filter.Limit > 0 {
            query += fmt.Sprintf(" LIMIT $%d", argPos)
            args = append(args, filter.Limit)
        }
    }

    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []*model.ChangelogEntry
    for rows.Next() {
        entry := &model.ChangelogEntry{}
        err := rows.Scan(
            &entry.ID,
            &entry.EntityType,
            &entry.EntityID,
            &entry.Operation,
            &entry.ChangedByUser,
            &entry.ChangedByClient,
            &entry.ChangedAt,
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
