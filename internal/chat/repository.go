package chat

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, m *Message) error
	GetByTicketID(ctx context.Context, ticketID int) ([]Message, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, m *Message) error {
	query := `INSERT INTO messages (ticket_id, sender_type, sender_name, message, is_attachment, attachment_type, attachment_url) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, m.TicketID, m.SenderType, m.SenderName, m.Message, m.IsAttachment, m.AttachmentType, m.AttachmentURL).Scan(
		&m.ID, &m.CreatedAt,
	)
}

func (r *repository) GetByTicketID(ctx context.Context, ticketID int) ([]Message, error) {
	query := `SELECT id, ticket_id, sender_type, sender_name, message, is_attachment, attachment_type, attachment_url, created_at 
	          FROM messages WHERE ticket_id = $1 ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("get messages error: %w", err)
	}
	defer rows.Close()

	var list []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderType, &m.SenderName, &m.Message, &m.IsAttachment, &m.AttachmentType, &m.AttachmentURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}
