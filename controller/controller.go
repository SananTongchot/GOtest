package controller

import (
    "encoding/json"
    "log"
    "net/http"
    "myapp/model"
    "myapp/config"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := config.DB.Query("SELECT id, name, email FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []model.User
    for rows.Next() {
        var user model.User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            log.Fatal(err)
        }
        users = append(users, user)
    }

    json.NewEncoder(w).Encode(users)
}
