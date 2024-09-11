package main

import (
	"log"
	"myapp/config" // Import the config package
	"myapp/router"
	"net/http"
)

func main() {
	// Initialize database connection using the config package
	config.Connect()

	// Initialize the router with the database connection from the config package
	r := router.InitRoutes(config.DB)

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
