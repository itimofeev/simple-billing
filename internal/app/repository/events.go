package repository

import (
	"time"

	"github.com/go-pg/pg/v10"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

func (r *Repository) AddEvent(tx pg.DBI, event model.Event) (model.Event, error) {
	_, err := tx.Model(&event).Returning("*").Insert()
	return event, err
}

func (r *Repository) ListEventsByFromUserID(tx pg.DBI, userID int64) (events []model.Event, err error) {
	return events, tx.Model(&events).Where("from_user_id = ?", userID).Select()
}

func (r *Repository) SetMessageID(tx pg.DBI, event model.Event, messageID string) error {
	_, err := tx.Model(&event).Set("queue_id = ?", messageID).WherePK().Update()
	return err
}

func (r *Repository) SetMessageSent(tx pg.DBI, messageID string, now time.Time) error {
	_, err := tx.Model(&model.Event{}).
		Set("queue_sent_time = ?", now).
		Where("queue_id = ?", messageID).
		Update()
	return err
}
