package router

import (
	"database/sql"
	controller "myapp/API"

	"github.com/gorilla/mux"
)

// InitRoutes initializes and returns the router with all the routes defined
func InitRoutes(db *sql.DB) *mux.Router {
	router := mux.NewRouter()

	// Routes for authentication
	router.HandleFunc("/register", controller.RegisterUser).Methods("POST")
	router.HandleFunc("/login", controller.LoginUser).Methods("POST")

	// Inject the database connection into the GenerateLotteryHandler
	router.HandleFunc("/random", controller.GenerateLotteryHandler(db)).Methods("POST")

	return router
}
