package router

import (
    "github.com/gorilla/mux"
    "myapp/controller"
)

func InitRoutes() *mux.Router {
    router := mux.NewRouter()

    router.HandleFunc("/users", controller.GetUsers).Methods("GET")

    return router
}
