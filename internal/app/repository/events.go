package repository

import (
	"github.com/go-pg/pg/v10"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

func (r *Repository) AddEvent(tx pg.DBI, event model.Event) error {
	_, err := tx.Model(&event).Insert()
	return err
}

func (r *Repository) ListEventsByFromUserID(tx pg.DBI, userID int64) (events []model.Event, err error) {
	return events, tx.Model(&events).Where("from_user_id = ?", userID).Select()
}
