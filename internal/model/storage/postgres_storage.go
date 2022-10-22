package storage

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/now"
	"log"

	// postgres driver
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"max.ks1230/project-base/internal/model/customerr"
	"time"
)

const dsnTemplate = "user=%s password=%s host=%s dbname=%s sslmode=disable"

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type config interface {
	Host() string
	Username() string
	Password() string
	Database() string
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(config config) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", fmt.Sprintf(dsnTemplate,
		config.Username(),
		config.Password(),
		config.Host(),
		config.Database()))
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to database")
	}
	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "cannot connect to database")
	}
	return &PostgresStorage{db}, nil
}

func (s *PostgresStorage) GetUserByID(ctx context.Context, id int64) (user.Record, error) {
	query := psql.Select("preferred_currency", "month_limit").
		From("users").
		Where(sq.Eq{"id": id})

	var res user.Record
	var curr string
	err := query.RunWith(s.db).QueryRowContext(ctx).Scan(&curr, &res.MonthLimit)
	if err != nil {
		return user.Record{}, errors.Wrap(err, "get user")
	}
	res.SetPreferredCurrency(curr)
	return res, nil
}

func (s *PostgresStorage) SaveUserByID(ctx context.Context, id int64, rec user.Record) error {
	query := psql.Insert("users").
		Columns("id", "preferred_currency", "month_limit", "updated_at").
		Values(id, rec.PreferredCurrency(), rec.MonthLimit, time.Now()).
		Suffix("ON CONFLICT(id) DO UPDATE SET preferred_currency = ?, month_limit = ?, updated_at = ?",
			rec.PreferredCurrency(), rec.MonthLimit, time.Now())

	_, err := query.RunWith(s.db).ExecContext(ctx)
	return errors.Wrap(err, "save user")
}

func (s *PostgresStorage) SaveExpense(ctx context.Context, userID int64, rec user.ExpenseRecord) error {
	query := psql.Insert("expenses").
		Columns("user_id", "amount", "category", "created_at").
		Values(userID, rec.Amount, rec.Category, rec.Created)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "save expense")
	}
	defer func() {
		txErr := tx.Rollback()
		if txErr != nil {
			log.Println("error when transaction rollback", txErr)
		}
	}()

	_, err = query.RunWith(tx).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "save expense")
	}
	limMet, err := s.isLimitMet(ctx, tx, userID)
	if err != nil {
		return errors.Wrap(err, "save expense")
	}
	if !limMet {
		return &customerr.LimitError{Err: "user limit exceeded"}
	}
	err = tx.Commit()
	return err
}

func (s *PostgresStorage) isLimitMet(ctx context.Context, tx *sql.Tx, userID int64) (bool, error) {
	query := `
	SELECT total.s <= total.lim OR total.lim = 0 AS test FROM
		(
			SELECT sum(e.amount) AS s, u.month_limit AS lim FROM expenses e
			JOIN users u ON u.id = e.user_id
			WHERE e.user_id = $1 AND e.created_at > $2 AND e.created_at < $3
			GROUP BY u.month_limit
		) AS total
`
	var test bool
	err := tx.QueryRowContext(ctx, query,
		userID, now.BeginningOfMonth(), now.EndOfMonth()).
		Scan(&test)
	if err != nil {
		return false, errors.Wrap(err, "ensure limit")
	}
	return test, nil
}

func (s *PostgresStorage) GetUserExpenses(ctx context.Context, userID int64) ([]user.ExpenseRecord, error) {
	query := psql.Select("amount", "category", "created_at").
		From("expenses").
		Where(sq.Eq{"user_id": userID})

	rows, err := query.RunWith(s.db).QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get expenses")
	}
	defer func() {
		rowErr := rows.Close()
		if rowErr != nil {
			log.Println("error closing rows", rowErr)
		}
	}()

	exps := make([]user.ExpenseRecord, 0)
	for rows.Next() {
		var e user.ExpenseRecord
		err = rows.Scan(&e.Amount, &e.Category, &e.Created)
		if err != nil {
			return nil, errors.Wrap(err, "get expenses")
		}
		exps = append(exps, e)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "get expenses")
	}

	return exps, nil
}

func (s *PostgresStorage) GetRate(ctx context.Context, name string) (currency.Rate, error) {
	query := psql.Select("name", "base_rate", "is_set", "updated_at").
		From("rates").
		Where(sq.Eq{"name": name}).
		OrderBy("updated_at DESC").
		Limit(1)

	var res currency.Rate
	err := query.RunWith(s.db).QueryRowContext(ctx).Scan(&res.Name, &res.BaseRate, &res.Set, &res.UpdatedAt)
	if err != nil {
		return currency.Rate{}, err
	}
	if !res.Set {
		return currency.Rate{}, fmt.Errorf("rate %s is not set yet", name)
	}
	return res, nil
}

func (s *PostgresStorage) NewRate(ctx context.Context, name string) error {
	query := psql.Insert("rates").
		Columns("name", "base_rate", "is_set").
		Values(name, 0, false)
	_, err := query.RunWith(s.db).ExecContext(ctx)
	return errors.Wrap(err, "new rate")
}

func (s *PostgresStorage) UpdateRateValue(ctx context.Context, name string, val float64) error {
	query := psql.Insert("rates").
		Columns("name", "base_rate", "is_set").
		Values(name, val, true)
	_, err := query.RunWith(s.db).ExecContext(ctx)
	return errors.Wrap(err, "update rate")
}
