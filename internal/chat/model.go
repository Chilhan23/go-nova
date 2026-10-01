package chat

import "time"

type Message struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	TicketID       int       `json:"ticket_id" gorm:"not null;index"`
	SenderType     string    `json:"sender_type" gorm:"type:varchar(20);not null"` // user, ai, programmer, system
	SenderName     *string   `json:"sender_name,omitempty" gorm:"type:varchar(150)"`
	Message        string    `json:"message" gorm:"type:text;not null"`
	IsAttachment   bool      `json:"is_attachment" gorm:"default:false"`
	AttachmentType *string   `json:"attachment_type,omitempty" gorm:"type:varchar(20)"`
	AttachmentURL  *string   `json:"attachment_url,omitempty" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}
