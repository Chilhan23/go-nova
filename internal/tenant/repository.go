package tenant

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	GetByKey(ctx context.Context, key string) (*Tenant, error)
	Create(ctx context.Context, t *Tenant) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByKey(ctx context.Context, key string) (*Tenant, error) {
	query := `SELECT id, key_identifier, app_name, tenant_name, api_key, is_active, created_at, updated_at 
	          FROM tenants WHERE key_identifier = $1 LIMIT 1`
	var t Tenant
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&t.ID, &t.KeyIdentifier, &t.AppName, &t.TenantName, &t.APIKey, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant error: %w", err)
	}
	return &t, nil
}

func (r *repository) Create(ctx context.Context, t *Tenant) error {
	query := `INSERT INTO tenants (key_identifier, app_name, tenant_name, is_active) 
	          VALUES ($1, $2, $3, true) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, t.KeyIdentifier, t.AppName, t.TenantName).Scan(
		&t.ID, &t.CreatedAt, &t.UpdatedAt,
	)
}
