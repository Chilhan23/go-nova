package tenant

import "time"

type Tenant struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	KeyIdentifier string    `json:"key_identifier" gorm:"uniqueIndex;type:varchar(100);not null"`
	AppName       string    `json:"app_name" gorm:"type:varchar(100);not null"`
	TenantName    string    `json:"tenant_name" gorm:"type:varchar(150);not null"`
	APIKey        *string   `json:"api_key,omitempty" gorm:"type:varchar(255)"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Tenant) TableName() string {
	return "tenants"
}
