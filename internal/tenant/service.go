package tenant

import (
	"context"
	"fmt"
)

type Service interface {
	GetOrCreate(ctx context.Context, req RegisterTenantRequest) (*Tenant, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetOrCreate(ctx context.Context, req RegisterTenantRequest) (*Tenant, error) {
	existing, err := s.repo.GetByKey(ctx, req.KeyIdentifier)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	newTenant := &Tenant{
		KeyIdentifier: req.KeyIdentifier,
		AppName:       req.AppName,
		TenantName:    req.TenantName,
	}

	if err := s.repo.Create(ctx, newTenant); err != nil {
		return nil, fmt.Errorf("create tenant failed: %w", err)
	}

	return newTenant, nil
}
