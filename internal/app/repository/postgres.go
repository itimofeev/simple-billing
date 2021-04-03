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
	db.AddQueryHook(dbLogger{})
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

func (r *Repository) GetBalance(tx pg.DBI, userID int64) (balance model.Balance, err error) {
	if err := tx.Model(&balance).Where("id = ?", userID).Select(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			err = model.ErrUserNotFound
		}
		return model.Balance{}, fmt.Errorf("[postgres] error on getting balance: %w", err)
	}
	return balance, nil
}

func (r *Repository) UpdateBalance(ctx context.Context, userID, amount int64) (balance model.Balance, err error) {
	// nolint
	//	_, err := s.db.Model(&app).WherePK().Set("history_time = now()").Update()
	if err := r.db.WithContext(ctx).Model(&balance).Where("id = ?", userID).Select(); err != nil {
		return model.Balance{}, fmt.Errorf("[postgres] error on getting balance: %w", err)
	}
	return balance, nil
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
