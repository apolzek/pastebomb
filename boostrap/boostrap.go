package boostrap

import (
	"gin-goinc-api/config"
	"gin-goinc-api/config/app_config"
	"gin-goinc-api/config/cors_config"
	"gin-goinc-api/database"
	"gin-goinc-api/router"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func BootstrapApp() {

	//load env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	//init config
	config.InitConfig()

	//database connection
	database.ConnectDatabase()

	//init gin engine
	app := gin.Default()

	//cors config
	app.Use(cors_config.CorsConfigContrib())
	// app.Use(cors_config.CorsConfig())

	//init router
	router.InitRoute(app)

	//run app
	app.Run(app_config.PORT)
}
