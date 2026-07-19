package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghstouch/one-go/internal/config"
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/pkg/logger"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Init initializes the database connection
func Init(cfg *config.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
		logger.Info("Connecting to PostgreSQL database...")
	case "sqlite":
		// Ensure directory exists for SQLite
		dir := filepath.Dir(cfg.SQLitePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
		dialector = sqlite.Open(cfg.SQLitePath)
		logger.Infof("Connecting to SQLite database at %s...", cfg.SQLitePath)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable WAL mode for SQLite for better concurrency
	if cfg.Driver == "sqlite" {
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA foreign_keys=ON")
		db.Exec("PRAGMA busy_timeout=5000")
	}

	logger.Info("Database connection established")
	return nil
}

// Get returns the database instance
func Get() *gorm.DB {
	return db
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	logger.Info("Running database migrations...")

	err := db.AutoMigrate(
		&model.User{},
		&model.Settings{},
		&model.Provider{},
		&model.Combo{},
		&model.ComboTarget{},
		&model.ApiKey{},
		&model.UsageHistory{},
		&model.CallLog{},
		&model.Proxy{},
		&model.ProxyLog{},
		&model.ModelAlias{},
		&model.AuditLog{},
		&model.Webhook{},
		&model.WebhookLog{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed")
	return nil
}

// Seed seeds initial data
func Seed(cfg *config.Config) error {
	logger.Info("Seeding initial data...")

	// Seed default admin user
	if err := seedAdminUser(cfg); err != nil {
		return err
	}

	// Seed default settings
	if err := seedDefaultSettings(); err != nil {
		return err
	}

	logger.Info("Initial data seeded")
	return nil
}

// seedAdminUser creates default admin user if not exists
func seedAdminUser(cfg *config.Config) error {
	var count int64
	db.Model(&model.User{}).Count(&count)

	if count == 0 {
		hashedPassword, err := hashPassword(cfg.Admin.Password)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		admin := model.User{
			Username: cfg.Admin.Username,
			Password: hashedPassword,
			Role:     model.RoleAdmin,
			IsActive: true,
		}

		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		logger.Infof("Default admin user created: %s", cfg.Admin.Username)
	}

	return nil
}

// seedDefaultSettings creates default settings if not exists
func seedDefaultSettings() error {
	defaults := model.DefaultSettings()

	for _, setting := range defaults {
		var existing model.Settings
		result := db.Where("key = ?", setting.Key).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&setting).Error; err != nil {
				return fmt.Errorf("failed to create setting %s: %w", setting.Key, err)
			}
		}
	}

	return nil
}

// Close closes the database connection
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
