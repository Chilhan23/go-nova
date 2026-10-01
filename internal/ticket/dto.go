package ticket

import (
	"encoding/json"
	"time"
)

type InitTicketRequest struct {
	KeyIdentifier  string          `json:"key_identifier" binding:"required"`
	AppName        string          `json:"app_name" binding:"required"`
	TenantName     string          `json:"tenant_name" binding:"required"`
	UserID         string          `json:"user_id"`
	UserName       string          `json:"user_name" binding:"required"`
	ModuleName     string          `json:"module_name"`
	DiagnosticInfo json.RawMessage `json:"diagnostic_info"`
}

type TicketResponse struct {
	ID                 int             `json:"id"`
	TicketCode         string          `json:"ticket_code"`
	UserName           string          `json:"user_name"`
	ModuleName         string          `json:"module_name"`
	TopicTitle         string          `json:"topic_title,omitempty"`
	AssignedProgrammer string          `json:"assigned_programmer,omitempty"`
	Status             string          `json:"status"`
	DiagnosticInfo     json.RawMessage `json:"diagnostic_info,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

type RateTicketRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review"`
}
