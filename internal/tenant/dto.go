package tenant

type RegisterTenantRequest struct {
	KeyIdentifier string `json:"key_identifier" binding:"required"`
	AppName       string `json:"app_name" binding:"required"`
	TenantName    string `json:"tenant_name" binding:"required"`
}

type TenantResponse struct {
	ID            int    `json:"id"`
	KeyIdentifier string `json:"key_identifier"`
	AppName       string `json:"app_name"`
	TenantName    string `json:"tenant_name"`
	IsActive      bool   `json:"is_active"`
}
