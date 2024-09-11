package myapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"myapp/config"
	"myapp/model"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUser handles user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate password
	if user.Password == "" {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert into the database
	_, err = config.DB.Exec("INSERT INTO user (username, phone, email, password, type, credit) VALUES (?, ?, ?, ?, ?, ?)",
		user.Username, user.Phone, user.Email, hashedPassword, "2", 10000)
	if err != nil {
		log.Println("Error inserting user:", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginUser handles user login
// LoginUser handles user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch the user from the database
	var user model.User
	err := config.DB.QueryRow("SELECT uid, password FROM user WHERE email = ?", credentials.Email).Scan(&user.UID, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			log.Println("Error fetching user:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Compare the passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate a simple response for now (you might want to use JWT for token-based authentication)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}


func generateRandomNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // สุ่มเลข 6 หลักระหว่าง 000000 ถึง 999999 และแปลงเป็น string
}

// ฟังก์ชัน handler สำหรับ endpoint /generate-lottery
// GenerateLotteryHandler returns an HTTP handler function that generates lottery numbers
func GenerateLotteryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		const numNumbers = 100
		lottoNumbers := make([]string, 0, numNumbers)
		uniqueNumbers := make(map[string]bool)

		// Generate unique lottery numbers
		for len(lottoNumbers) < numNumbers {
			lottoNumber := generateRandomNumber()
			if _, exists := uniqueNumbers[lottoNumber]; !exists {
				uniqueNumbers[lottoNumber] = true
				lottoNumbers = append(lottoNumbers, lottoNumber)
			}
		}

		// Insert the numbers into the database
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}

		defer func() {
			if err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				tx.Commit()
				response := model.LotteryResponse{
					Message: fmt.Sprintf("%d lottery numbers generated and saved successfully", numNumbers),
					Number:  "",
				}
				json.NewEncoder(w).Encode(response)
			}
		}()

		for _, lottoNumber := range lottoNumbers {
			_, err := tx.Exec("INSERT INTO lottery (lotto_number) VALUES (?)", lottoNumber)
			if err != nil {
				return // return immediately to trigger rollback
			}
		}
	}
}
