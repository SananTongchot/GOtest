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
	router.HandleFunc("/", controller.Test).Methods("GET")
	// Inject the database connection into the GenerateLotteryHandler
	router.HandleFunc("/random", controller.GenerateLotteryHandler(db)).Methods("POST")

	router.HandleFunc("/buy_lottery", controller.BuyLottery).Methods("POST")
	return router
}
