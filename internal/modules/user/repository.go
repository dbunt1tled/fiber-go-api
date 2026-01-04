package user

import (
	"database/sql"

	"github.com/dbunt1tled/fiber-go-api/pkg/storage"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	*storage.Repository[*User]
}

func NewUserRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		Repository: storage.NewRepository[*User](
			pool,
			"users",
			scanUser,
			scanUsers,
			buildUserRecord,
		),
	}
}

func scanUser(row pgx.Row) (*User, error) {
	var user User
	var confirmedAt sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.SecondName,
		&user.Email,
		&user.PhoneNumber,
		&user.Status,
		&user.Password,
		&user.Roles,
		&user.Address,
		&confirmedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if confirmedAt.Valid {
		user.ConfirmedAt = &confirmedAt.Time
	}

	return &user, nil
}

func scanUsers(rows pgx.Rows) ([]*User, error) {
	return storage.ScanRowsWithScanner(rows, scanUser)
}

func buildUserRecord(user *User) goqu.Record {
	record := goqu.Record{
		"id":           user.ID,
		"first_name":   user.FirstName,
		"second_name":  user.SecondName,
		"email":        user.Email,
		"phone_number": user.PhoneNumber,
		"status":       user.Status,
		"password":     user.Password,
		"roles":        storage.ToPgArray(user.Roles),
		"address":      user.Address,
		"updated_at":   goqu.L("NOW()"),
		"created_at":   user.CreatedAt,
	}

	if user.ConfirmedAt != nil {
		record["confirmed_at"] = user.ConfirmedAt
	}

	return record
}
