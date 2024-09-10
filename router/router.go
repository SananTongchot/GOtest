package router

import (
	"myapp/controller"
	"github.com/gorilla/mux"
)

// InitRoutes initializes and returns the router with all the routes defined
func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Routes for authentication
	router.HandleFunc("/register", controller.RegisterUser).Methods("POST")
	router.HandleFunc("/login", controller.LoginUser).Methods("POST")

	return router
}
