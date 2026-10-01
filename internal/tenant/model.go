package tenant

import "time"

type Tenant struct {
	ID            int       `json:"id"`
	KeyIdentifier string    `json:"key_identifier"`
	AppName       string    `json:"app_name"`
	TenantName    string    `json:"tenant_name"`
	APIKey        *string   `json:"api_key,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
