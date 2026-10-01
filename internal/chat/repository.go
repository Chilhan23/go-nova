package chat

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *Message) error
	GetByTicketID(ctx context.Context, ticketID int) ([]Message, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, m *Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *repository) GetByTicketID(ctx context.Context, ticketID int) ([]Message, error) {
	var list []Message
	err := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("id ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
