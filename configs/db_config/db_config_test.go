package db_config

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestInitDatabaseConfig(t *testing.T) {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Check envs
	if os.Getenv("DB_DRIVER") == "" {
		t.Error("Environment variable DB_DRIVER is not defined.")
	}
	if os.Getenv("DB_HOST") == "" {
		t.Error("Environment variable DB_HOST is not defined.")
	}
	if os.Getenv("DB_PORT") == "" {
		t.Error("Environment variable DB_PORT is not defined.")
	}
	if os.Getenv("DB_NAME") == "" {
		t.Error("Environment variable DB_NAME is not defined.")
	}
	if os.Getenv("DB_USER") == "" {
		t.Error("Environment variable DB_USER is not defined.")
	}
	if os.Getenv("DB_PASSOWRD") == "" {
		t.Error("Environment variable DB_PASSOWRD is not defined.")
	}
}
