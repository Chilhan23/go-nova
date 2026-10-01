package chat

import "time"

type SendMessageRequest struct {
	TicketID int    `json:"ticket_id" binding:"required"`
	Message  string `json:"message" binding:"required"`
}

type SendMessageResponse struct {
	Status    bool   `json:"status"`
	UserMsgID int    `json:"user_msg_id"`
	AIMsgID   *int   `json:"ai_msg_id,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Reply     string `json:"reply,omitempty"`
}

type UploadAttachmentResponse struct {
	Status    bool   `json:"status"`
	UserMsgID int    `json:"user_msg_id"`
	AIMsgID   *int   `json:"ai_msg_id,omitempty"`
	Sender    string `json:"sender,omitempty"`
	FileURL   string `json:"file_url"`
	Reply     string `json:"reply,omitempty"`
}

type MessageItemResponse struct {
	ID             int       `json:"id"`
	TicketID       int       `json:"ticket_id"`
	SenderType     string    `json:"sender_type"`
	SenderName     string    `json:"sender_name,omitempty"`
	Message        string    `json:"message"`
	IsAttachment   bool      `json:"is_attachment"`
	AttachmentType string    `json:"attachment_type,omitempty"`
	AttachmentURL  string    `json:"attachment_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
