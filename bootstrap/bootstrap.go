package bootstrap

import (
	"fmt"
	"log"
	"time"

	"gin-goinc-api/config"
	"gin-goinc-api/config/app_config"
	"gin-goinc-api/config/cors_config"
	"gin-goinc-api/database"
	"gin-goinc-api/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func BootstrapApp() error {
	// Loading .env is best-effort: in containers the env is pre-populated.
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file loaded (this is fine in containers)")
	}

	config.InitConfig()

	if err := database.ConnectDatabase(); err != nil {
		return fmt.Errorf("database: %w", err)
	}

	app := gin.Default()
	app.Use(cors_config.CorsConfigContrib())
	router.InitRoute(app)

	// Sweep expired pastes once a minute. Adjust if traffic grows.
	startExpiredPostPurger(1 * time.Minute)

	return app.Run(app_config.PORT)
}
