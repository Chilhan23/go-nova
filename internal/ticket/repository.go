package ticket

import (
	"context"
	"database/sql"
	"fmt"
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
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetActiveTicket(ctx context.Context, tenantID int, userID string) (*Ticket, error) {
	query := `SELECT id, tenant_id, ticket_code, user_id, user_name, module_name, topic_title, 
	                 telegram_thread_id, assigned_programmer, status, csat_rating, csat_review, 
	                 COALESCE(diagnostic_info, '{}'::jsonb), created_at, updated_at 
	          FROM tickets 
	          WHERE tenant_id = $1 AND status != 'resolved' `
	var args []interface{}
	args = append(args, tenantID)

	if userID != "" {
		query += ` AND user_id = $2 `
		args = append(args, userID)
	}
	query += ` ORDER BY id DESC LIMIT 1`

	var t Ticket
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID, &t.TenantID, &t.TicketCode, &t.UserID, &t.UserName, &t.ModuleName, &t.TopicTitle,
		&t.TelegramThreadID, &t.AssignedProgrammer, &t.Status, &t.CSATRating, &t.CSATReview,
		&t.DiagnosticInfo, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active ticket error: %w", err)
	}
	return &t, nil
}

func (r *repository) GetByID(ctx context.Context, id int) (*Ticket, error) {
	query := `SELECT id, tenant_id, ticket_code, user_id, user_name, module_name, topic_title, 
	                 telegram_thread_id, assigned_programmer, status, csat_rating, csat_review, 
	                 COALESCE(diagnostic_info, '{}'::jsonb), created_at, updated_at 
	          FROM tickets WHERE id = $1 LIMIT 1`
	var t Ticket
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.TenantID, &t.TicketCode, &t.UserID, &t.UserName, &t.ModuleName, &t.TopicTitle,
		&t.TelegramThreadID, &t.AssignedProgrammer, &t.Status, &t.CSATRating, &t.CSATReview,
		&t.DiagnosticInfo, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get ticket by id error: %w", err)
	}
	return &t, nil
}

func (r *repository) GetByThreadID(ctx context.Context, threadID int64) (*Ticket, error) {
	query := `SELECT id, tenant_id, ticket_code, user_id, user_name, module_name, topic_title, 
	                 telegram_thread_id, assigned_programmer, status, csat_rating, csat_review, 
	                 COALESCE(diagnostic_info, '{}'::jsonb), created_at, updated_at 
	          FROM tickets WHERE telegram_thread_id = $1 AND status != 'resolved' LIMIT 1`
	var t Ticket
	err := r.db.QueryRowContext(ctx, query, threadID).Scan(
		&t.ID, &t.TenantID, &t.TicketCode, &t.UserID, &t.UserName, &t.ModuleName, &t.TopicTitle,
		&t.TelegramThreadID, &t.AssignedProgrammer, &t.Status, &t.CSATRating, &t.CSATReview,
		&t.DiagnosticInfo, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get ticket by thread id error: %w", err)
	}
	return &t, nil
}

func (r *repository) Create(ctx context.Context, t *Ticket) error {
	diag := t.DiagnosticInfo
	if len(diag) == 0 {
		diag = []byte("{}")
	}
	query := `INSERT INTO tickets (tenant_id, ticket_code, user_id, user_name, module_name, diagnostic_info, status) 
	          VALUES ($1, $2, $3, $4, $5, $6, 'open') RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, t.TenantID, t.TicketCode, t.UserID, t.UserName, t.ModuleName, diag).Scan(
		&t.ID, &t.CreatedAt, &t.UpdatedAt,
	)
}

func (r *repository) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *repository) UpdateThreadID(ctx context.Context, id int, threadID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tickets SET telegram_thread_id = $1, updated_at = NOW() WHERE id = $2`, threadID, id)
	return err
}

func (r *repository) UpdateTopicTitle(ctx context.Context, id int, title string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE tickets SET topic_title = $1, updated_at = NOW() WHERE id = $2`, title, id)
	return err
}

func (r *repository) UpdateProgrammer(ctx context.Context, id int, name string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tickets 
		SET assigned_programmer = $1, status = 'escalated', updated_at = NOW() 
		WHERE id = $2 AND (assigned_programmer IS NULL OR assigned_programmer = '')`, name, id)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (r *repository) SaveRating(ctx context.Context, id int, rating int, review string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tickets 
		SET csat_rating = $1, csat_review = $2, status = 'resolved', updated_at = NOW() 
		WHERE id = $3`, rating, review, id)
	return err
}
