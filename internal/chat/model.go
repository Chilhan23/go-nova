package chat

import "time"

type Message struct {
	ID             int       `json:"id"`
	TicketID       int       `json:"ticket_id"`
	SenderType     string    `json:"sender_type"` // user, ai, programmer, system
	SenderName     *string   `json:"sender_name,omitempty"`
	Message        string    `json:"message"`
	IsAttachment   bool      `json:"is_attachment"`
	AttachmentType *string   `json:"attachment_type,omitempty"` // image, video, document
	AttachmentURL  *string   `json:"attachment_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
