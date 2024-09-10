// package main

// import (
// 	"log"
// 	"net/http"
// 	"myapp/controller"
// 	"myapp/config"
// 	"github.com/gorilla/mux"
// )

// func main() {
// 	config.Connect()

// 	r := mux.NewRouter()

// 	r.HandleFunc("/register", controller.RegisterUser).Methods("POST")
// 	r.HandleFunc("/login", controller.LoginUser).Methods("POST")

//		http.Handle("/", r)
//		log.Fatal(http.ListenAndServe(":8080", nil))
//	}
package main

import (
	"log"
	"myapp/config"
	"myapp/router"
	"net/http"
)

func main() {
	// Connect to the database
	config.Connect()

	// Initialize routes
	r := router.InitRoutes()

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
