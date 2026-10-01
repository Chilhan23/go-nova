package ticket

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	GetActiveTicket(ctx context.Context, tenantID int, userID string) (*Ticket, error)
	GetByID(ctx context.Context, id int) (*Ticket, error)
	GetByThreadID(ctx context.Context, threadID int64) (*Ticket, error)
	Create(ctx context.Context, t *Ticket) error
	UpdateStatus(ctx context.Context, id int, status string) error
	UpdateThreadID(ctx context.Context, id int, threadID int64) error
	UpdateTopicTitle(ctx context.Context, id int, title string) error
	UpdateProgrammer(ctx context.Context, id int, name string) (bool, error)
	SaveRating(ctx context.Context, id int, rating int, review string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetActiveTicket(ctx context.Context, tenantID int, userID string) (*Ticket, error) {
	var t Ticket
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND status != 'resolved'", tenantID)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Order("id DESC").First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) GetByID(ctx context.Context, id int) (*Ticket, error) {
	var t Ticket
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) GetByThreadID(ctx context.Context, threadID int64) (*Ticket, error) {
	var t Ticket
	err := r.db.WithContext(ctx).Where("telegram_thread_id = ? AND status != 'resolved'", threadID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) Create(ctx context.Context, t *Ticket) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repository) UpdateStatus(ctx context.Context, id int, status string) error {
	return r.db.WithContext(ctx).Model(&Ticket{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) UpdateThreadID(ctx context.Context, id int, threadID int64) error {
	return r.db.WithContext(ctx).Model(&Ticket{}).Where("id = ?", id).Update("telegram_thread_id", threadID).Error
}

func (r *repository) UpdateTopicTitle(ctx context.Context, id int, title string) error {
	return r.db.WithContext(ctx).Model(&Ticket{}).Where("id = ?", id).Update("topic_title", title).Error
}

func (r *repository) UpdateProgrammer(ctx context.Context, id int, name string) (bool, error) {
	res := r.db.WithContext(ctx).Model(&Ticket{}).
		Where("id = ? AND (assigned_programmer IS NULL OR assigned_programmer = '')", id).
		Updates(map[string]interface{}{
			"assigned_programmer": name,
			"status":              "escalated",
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *repository) SaveRating(ctx context.Context, id int, rating int, review string) error {
	return r.db.WithContext(ctx).Model(&Ticket{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"csat_rating": rating,
			"csat_review": review,
			"status":      "resolved",
		}).Error
}
