package database

import (
	"time"
)

// CAInfo stores Certificate Authority information
type CAInfo struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null;unique" json:"name"`
	Organization string    `gorm:"not null" json:"organization"`
	Country      string    `gorm:"not null" json:"country"`
	KeyHash      string    `gorm:"not null" json:"key_hash"` // Hash of the CA key for verification
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Certificate stores certificate information
type Certificate struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"not null;unique" json:"name"`
	SerialNumber string     `gorm:"not null;unique" json:"serial_number"`
	Subject      string     `gorm:"not null" json:"subject"`
	Issuer       string     `gorm:"not null" json:"issuer"`
	NotBefore    time.Time  `json:"not_before"`
	NotAfter     time.Time  `json:"not_after"`
	IsRevoked    bool       `gorm:"default:false" json:"is_revoked"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	S3Path       string     `json:"s3_path,omitempty"` // Path in S3 bucket
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// EmailSettings stores email configuration
type EmailSettings struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SMTPServer  string    `gorm:"not null" json:"smtp_server"`
	SMTPPort    string    `gorm:"not null" json:"smtp_port"`
	Username    string    `gorm:"not null" json:"username"`
	Password    string    `gorm:"not null" json:"password"` // Should be encrypted
	FromEmail   string    `gorm:"not null" json:"from_email"`
	ToEmail     string    `gorm:"not null" json:"to_email"`
	UseTLS      bool      `gorm:"default:false" json:"use_tls"`
	UseStartTLS bool      `gorm:"default:false" json:"use_start_tls"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuditLog stores audit trail for all operations
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Action     string    `gorm:"not null" json:"action"`
	Resource   string    `gorm:"not null" json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	UserIP     string    `json:"user_ip,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Details    string    `gorm:"type:text" json:"details,omitempty"`
	Success    bool      `gorm:"default:true" json:"success"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// SerialMapping stores the mapping between serial numbers and certificate names
type SerialMapping struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SerialNumber string    `gorm:"not null;unique" json:"serial_number"`
	CertName     string    `gorm:"not null" json:"cert_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName methods for custom table names
func (CAInfo) TableName() string {
	return "ca_info"
}

func (Certificate) TableName() string {
	return "certificates"
}

func (EmailSettings) TableName() string {
	return "email_settings"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func (SerialMapping) TableName() string {
	return "serial_mappings"
}
