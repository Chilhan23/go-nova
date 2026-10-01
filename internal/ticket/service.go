package ticket

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"tech-nova/internal/tenant"
)

type Service interface {
	GetOrCreateActive(ctx context.Context, req InitTicketRequest, tenant *tenant.Tenant) (*Ticket, error)
	GetByID(ctx context.Context, id int) (*Ticket, error)
	Rate(ctx context.Context, ticketID int, req RateTicketRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetOrCreateActive(ctx context.Context, req InitTicketRequest, t *tenant.Tenant) (*Ticket, error) {
	active, err := s.repo.GetActiveTicket(ctx, t.ID, req.UserID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return active, nil
	}

	code := fmt.Sprintf("TCK-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:6])
	newTicket := &Ticket{
		TenantID:   t.ID,
		TicketCode: code,
		UserName:   req.UserName,
		ModuleName: req.ModuleName,
		Status:     "open",
	}
	if len(req.DiagnosticInfo) > 0 {
		diagStr := string(req.DiagnosticInfo)
		newTicket.DiagnosticInfo = &diagStr
	}
	if req.UserID != "" {
		newTicket.UserID = &req.UserID
	}

	if err := s.repo.Create(ctx, newTicket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return newTicket, nil
}

func (s *service) GetByID(ctx context.Context, id int) (*Ticket, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) Rate(ctx context.Context, ticketID int, req RateTicketRequest) error {
	return s.repo.SaveRating(ctx, ticketID, req.Rating, req.Review)
}
