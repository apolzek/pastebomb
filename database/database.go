package database

import (
	"fmt"
	"gin-goinc-api/config/db_config"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var errConnection error

func ConnectDatabase() {

	for i := 1; i <= 3; i++ { // Defina o número máximo de tentativas (neste exemplo, 3)

		if db_config.DB_DRIVER == "test" {
			dsnMysql := "user:1234@tcp(localhost:3306)/go_gin_gonic?charset=utf8mb4&parseTime=True&loc=Local"
			DB, errConnection = gorm.Open(mysql.Open(dsnMysql), &gorm.Config{})
		}

		if db_config.DB_DRIVER == "mysql" {
			dsnMysql := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db_config.DB_USER, db_config.DB_PASSWORD, db_config.DB_HOST, db_config.DB_PORT, db_config.DB_NAME)
			DB, errConnection = gorm.Open(mysql.Open(dsnMysql), &gorm.Config{})
		}

		if db_config.DB_DRIVER == "pgsql" {
			dsnMysql := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta", db_config.DB_HOST, db_config.DB_USER, db_config.DB_PASSWORD, db_config.DB_NAME, db_config.DB_PORT)
			DB, errConnection = gorm.Open(postgres.Open(dsnMysql), &gorm.Config{})
		}

		if errConnection == nil {
			log.Println("Connect to database")
			return // Se a conexão for bem-sucedida, saia do loop
		}

		log.Printf("Failed to connect to database (attempt %d): %v\n", i, errConnection)

		// Aguarde 2 segundos antes da próxima tentativa
		time.Sleep(5 * time.Second)
	}

	// Se todas as tentativas falharem, gere um erro fatal
	panic("Failed to connect to database after multiple attempts")
}
