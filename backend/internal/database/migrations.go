package database

import (
	"log"

	"github.com/nakamura/chatwoot-go/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Account{},
		&models.User{},
		&models.AccountUser{},
		&models.Inbox{},
		&models.Contact{},
		&models.Conversation{},
		&models.Message{},
		&models.Attachment{},
		&models.Team{},
		&models.Label{},
		&models.Webhook{},
		&models.AccessToken{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
