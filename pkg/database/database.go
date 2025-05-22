package database

import (
	"fmt"
	"time"

	"github.com/Lazarev-Cloud/localca-go/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	if !cfg.DatabaseEnabled {
		return nil, fmt.Errorf("database is not enabled")
	}

	// Build connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
		cfg.DatabasePort,
		cfg.DatabaseSSLMode,
	)

	// Configure GORM logger based on log level
	var gormLogger logger.Interface
	switch cfg.LogLevel {
	case "debug":
		gormLogger = logger.Default.LogMode(logger.Info)
	case "info":
		gormLogger = logger.Default.LogMode(logger.Warn)
	default:
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{DB: db}, nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	return d.DB.AutoMigrate(
		&CAInfo{},
		&Certificate{},
		&EmailSettings{},
		&AuditLog{},
		&SerialMapping{},
	)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database connectivity
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// LogAudit logs an audit entry
func (d *Database) LogAudit(action, resource, resourceID, userIP, userAgent, details string, success bool, errorMsg string) error {
	audit := AuditLog{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		UserIP:     userIP,
		UserAgent:  userAgent,
		Details:    details,
		Success:    success,
		Error:      errorMsg,
		CreatedAt:  time.Now().UTC(),
	}

	return d.DB.Create(&audit).Error
}

// GetAuditLogs retrieves audit logs with pagination
func (d *Database) GetAuditLogs(limit, offset int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	// Get total count
	if err := d.DB.Model(&AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := d.DB.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}
