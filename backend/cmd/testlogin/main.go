package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID           string `gorm:"primaryKey"`
	Email        string
	PasswordHash string
	Name         string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://chatwoot:chatwoot123@postgres:5432/chatwoot_go?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Find user
	var user User
	result := db.Table("users").Where("email = ?", "admin@nakamura.com").First(&user)
	if result.Error != nil {
		log.Fatal("User not found:", result.Error)
	}

	fmt.Printf("Found user: %s (%s)\n", user.Name, user.Email)
	fmt.Printf("Hash: %s\n", user.PasswordHash)
	fmt.Printf("Hash length: %d\n", len(user.PasswordHash))

	// Test password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("admin123"))
	if err != nil {
		fmt.Printf("Password check failed: %v\n", err)
	} else {
		fmt.Println("Password is correct!")
	}
}
