package tenant

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	GetByKey(ctx context.Context, key string) (*Tenant, error)
	Create(ctx context.Context, t *Tenant) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByKey(ctx context.Context, key string) (*Tenant, error) {
	var t Tenant
	err := r.db.WithContext(ctx).Where("key_identifier = ?", key).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) Create(ctx context.Context, t *Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}
