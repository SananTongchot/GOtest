package myapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"myapp/config"
	"myapp/model"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("Invalid input:", err)
		http.Error(w, `{"message": "Invalid input"}`, http.StatusBadRequest)
		return
	}

	// Validate password
	if user.Password == "" {
		http.Error(w, `{"message": "Password cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// Check if the email is already registered
	var existingUser string
	err := config.DB.QueryRow("SELECT email FROM user WHERE email = ?", user.Email).Scan(&existingUser)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error checking email existence:", err)
		http.Error(w, `{"message": "Failed to register user"}`, http.StatusInternalServerError)
		return
	}
	if existingUser != "" {
		http.Error(w, `{"message": "Email is already registered"}`, http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to hash password:", err)
		http.Error(w, `{"message": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Insert into the database
	_, err = config.DB.Exec("INSERT INTO user (username, phone, email, password, type) VALUES (?, ?, ?, ?, ?)",
		user.Username, user.Phone, user.Email, hashedPassword, "2")
	if err != nil {
		log.Println("Error inserting user:", err)
		http.Error(w, `{"message": "Failed to register user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginUser handles user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, `{"message": "Invalid input"}`, http.StatusBadRequest)
		return
	}

	var user model.User
	err := config.DB.QueryRow("SELECT uid, password, type FROM user WHERE email = ?", credentials.Email).Scan(&user.UID, &user.Password, &user.Type)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"message": "Invalid email or password"}`, http.StatusUnauthorized)
		} else {
			log.Println("Error fetching user:", err)
			http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, `{"message": "Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"uid":     user.UID,
		"type":    user.Type,
	})
}

// Test function for testing
func Test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Test Successful"})
}
