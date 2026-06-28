package database

import (
	"fmt"
	"log"
	"time"

	"gin-goinc-api/config/db_config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	maxConnectAttempts = 3
	connectBackoff     = 5 * time.Second
)

func ConnectDatabase() error {
	var lastErr error

	for i := 1; i <= maxConnectAttempts; i++ {
		db, err := openDB()
		if err == nil {
			DB = db
			log.Println("connected to database")
			return nil
		}

		lastErr = err
		log.Printf("failed to connect to database (attempt %d/%d): %v", i, maxConnectAttempts, err)

		if i < maxConnectAttempts {
			time.Sleep(connectBackoff)
		}
	}

	return fmt.Errorf("connect after %d attempts: %w", maxConnectAttempts, lastErr)
}

func openDB() (*gorm.DB, error) {
	switch db_config.DB_DRIVER {
	case "test", "mysql":
		user, pass, host, port, name := db_config.DB_USER, db_config.DB_PASSWORD, db_config.DB_HOST, db_config.DB_PORT, db_config.DB_NAME
		if db_config.DB_DRIVER == "test" {
			user, pass, host, port, name = "user", "1234", "localhost", "3306", "go_gin_gonic"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})

	case "pgsql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			db_config.DB_HOST, db_config.DB_USER, db_config.DB_PASSWORD, db_config.DB_NAME, db_config.DB_PORT)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unknown DB_DRIVER %q (want mysql|pgsql|test)", db_config.DB_DRIVER)
	}
}
