package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds PostgreSQL connection configuration
// Optimized for production environments with connection pooling
type Config struct {
	Host            string          `json:"host"`
	Port            int             `json:"port"`
	User            string          `json:"user"`
	Password        string          `json:"password"`
	Database        string          `json:"database"`
	SSLMode         string          `json:"ssl_mode"`           // disable, require, verify-ca, verify-full
	Timezone        string          `json:"timezone"`           // Default: UTC
	MaxIdleConns    int             `json:"max_idle_conns"`     // Default: 10
	MaxOpenConns    int             `json:"max_open_conns"`     // Default: 100
	ConnMaxLifetime time.Duration   `json:"conn_max_lifetime"`  // Default: 1 hour
	ConnMaxIdleTime time.Duration   `json:"conn_max_idle_time"` // Default: 30 minutes
	LogLevel        logger.LogLevel `json:"log_level"`          // Default: Silent in production
}

// NewConfig creates a new PostgreSQL configuration with production defaults
func NewConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		Database:        "postgres",
		SSLMode:         "disable",
		Timezone:        "UTC",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		LogLevel:        logger.Silent, // Production default
	}
}

// DSN builds the PostgreSQL connection string
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode, c.Timezone,
	)
}

// Connect establishes a connection to PostgreSQL with optimized settings
func Connect(config *Config) (*gorm.DB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(config.LogLevel),
		DisableForeignKeyConstraintWhenMigrating: false,
		CreateBatchSize:                          1000,  // Optimize batch operations
		PrepareStmt:                              true,  // Use prepared statements for better performance
		SkipDefaultTransaction:                   false, // Maintain ACID compliance
	}

	// Open connection
	db, err := gorm.Open(postgres.Open(config.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB instance: %w", err)
	}

	// Set connection pool parameters for optimal performance
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// MustConnect is like Connect but panics on error
// Useful for application startup where DB connectivity is critical
func MustConnect(config *Config) *gorm.DB {
	db, err := Connect(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to PostgreSQL: %v", err))
	}
	return db
}
