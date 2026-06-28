package config

import (
	"gin-goinc-api/config/app_config"
	"gin-goinc-api/config/db_config"
)

func InitConfig() {

	app_config.InitAppConfig()
	db_config.InitDatabaseConfig()

}
