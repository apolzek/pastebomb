package main

import (
	"log"

	"gin-goinc-api/bootstrap"
)

func main() {
	if err := bootstrap.BootstrapApp(); err != nil {
		log.Fatalf("startup failed: %v", err)
	}
}
