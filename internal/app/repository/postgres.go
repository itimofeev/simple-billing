package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-pg/pg/v10"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

type Repository struct {
	db *pg.DB
}

func New(url string) *Repository {
	opts, err := pg.ParseURL(url)
	if err != nil {
		panic(err)
	}
	db := pg.Connect(opts)
	db.AddQueryHook(dbLogger{enabled: false})
	if err := db.Ping(context.Background()); err != nil {
		panic(err)
	}

	if err := doMigrationIfNeeded(db); err != nil {
		panic(err)
	}

	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateUser(tx pg.DBI, userID int64) error {
	_, err := tx.Model(&model.Balance{
		UserID:  userID,
		Balance: 0,
	}).Insert()
	return err
}

func (r *Repository) GetBalance(tx pg.DBI, userID int64, withLock bool) (balance model.Balance, err error) {
	query := tx.Model(&balance).Where("id = ?", userID)
	if withLock {
		query = query.For("UPDATE")
	}
	if err := query.Select(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			err = model.ErrUserNotFound
		}
		return model.Balance{}, fmt.Errorf("[postgres] error on getting balance: %w", err)
	}
	return balance, nil
}

func (r *Repository) UpdateBalance(tx pg.DBI, userID, newBalance int64) error {
	balance := model.Balance{
		UserID: userID,
	}
	_, err := tx.Model(&balance).WherePK().Set("balance = ?", newBalance).Update()
	return err
}

func (r *Repository) DoInTX(ctx context.Context, f func(tx pg.DBI) error) error {
	return r.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return f(tx)
	})
}

func (r *Repository) GetDB(ctx context.Context) pg.DBI {
	return r.db.WithContext(ctx)
}

type dbLogger struct {
	enabled bool
}

func (d dbLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d dbLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	if d.enabled {
		query, _ := q.FormattedQuery()
		fmt.Println(string(query))
	}
	return nil
}
