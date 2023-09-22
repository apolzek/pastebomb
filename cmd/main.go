package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/apolzek/config"
	"github.com/apolzek/router"
)

func main() {
	config.LoadConfig()
	r := router.GenerateRoutes()

	fmt.Printf("Listening port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r))
}
