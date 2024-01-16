package app_config

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestInitAppConfig(t *testing.T) {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	envPort := os.Getenv("APP_PORT")
	envStaticRoute := os.Getenv("STATIC_ROUTE")
	envStaticDir := os.Getenv("STATIC_DIR")

	if envPort == "" {
		t.Error("Environment variable APP_PORT is not defined.")
	}

	if envStaticRoute == "" {
		t.Error("Environment variable STATIC_ROUTE is not defined.")
	}

	if envStaticDir == "" {
		t.Error("Environment variable STATIC_DIR is not defined.")
	}
}
