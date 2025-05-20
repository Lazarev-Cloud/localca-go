package handlers

// APIResponse is the standard response format for API calls
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// CertificateInfo represents certificate information for display
type CertificateInfo struct {
	CommonName     string `json:"common_name"`
	ExpiryDate     string `json:"expiry_date"`
	IsClient       bool   `json:"is_client"`
	SerialNumber   string `json:"serial_number"`
	IsExpired      bool   `json:"is_expired"`
	IsExpiringSoon bool   `json:"is_expiring_soon"`
	IsRevoked      bool   `json:"is_revoked"`
}

// CAInfo represents CA information for display
type CAInfo struct {
	CommonName   string `json:"common_name"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	ExpiryDate   string `json:"expiry_date"`
	IsExpired    bool   `json:"is_expired"`
}
