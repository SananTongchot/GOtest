package model

type User struct {
    UID       int    `json:"uid"`
    Username  string `json:"username"`
    Phone     string `json:"phone"`
    Email     string `json:"email"`
    Password  string `json:"password"`
    Type      string `json:"type"` // "1" for admin, "2" for user
    Credit    int    `json:"credit"`
}
