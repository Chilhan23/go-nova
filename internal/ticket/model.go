package ticket

import (
	"encoding/json"
	"time"
)

type Ticket struct {
	ID                 int             `json:"id"`
	TenantID           int             `json:"tenant_id"`
	TicketCode         string          `json:"ticket_code"`
	UserID             *string         `json:"user_id,omitempty"`
	UserName           string          `json:"user_name"`
	ModuleName         string          `json:"module_name"`
	TopicTitle         *string         `json:"topic_title,omitempty"`
	TelegramThreadID   *int64          `json:"telegram_thread_id,omitempty"`
	AssignedProgrammer *string         `json:"assigned_programmer,omitempty"`
	Status             string          `json:"status"` // open, escalated, waiting_user, resolved
	CSATRating         *int            `json:"csat_rating,omitempty"`
	CSATReview         *string         `json:"csat_review,omitempty"`
	DiagnosticInfo     json.RawMessage `json:"diagnostic_info,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
