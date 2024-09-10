package main

import (
    "log"
    "net/http"
    "myapp/config"
    "myapp/router"
)

func main() {
    config.Connect()

    r := router.InitRoutes()

    log.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
