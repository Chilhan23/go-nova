package ticket

import (
	"time"
)

type Ticket struct {
	ID                 int        `json:"id" gorm:"primaryKey"`
	TenantID           int        `json:"tenant_id" gorm:"not null;index"`
	TicketCode         string     `json:"ticket_code" gorm:"uniqueIndex;type:varchar(50);not null"`
	UserID             *string    `json:"user_id,omitempty" gorm:"type:varchar(100)"`
	UserName           string     `json:"user_name" gorm:"type:varchar(150);not null"`
	ModuleName         string     `json:"module_name" gorm:"type:varchar(100);default:'Umum'"`
	TopicTitle         *string    `json:"topic_title,omitempty" gorm:"type:varchar(255)"`
	TelegramThreadID   *int64     `json:"telegram_thread_id,omitempty" gorm:"index"`
	AssignedProgrammer *string    `json:"assigned_programmer,omitempty" gorm:"type:varchar(150)"`
	Status             string     `json:"status" gorm:"type:varchar(30);default:'open'"` // open, escalated, waiting_user, resolved
	CSATRating         *int       `json:"csat_rating,omitempty"`
	CSATReview         *string    `json:"csat_review,omitempty"`
	DiagnosticInfo     *string    `json:"diagnostic_info,omitempty" gorm:"type:jsonb"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (Ticket) TableName() string {
	return "tickets"
}
