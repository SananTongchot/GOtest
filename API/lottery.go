package myapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"myapp/model"
	"net/http"
	"time"
)

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

// generateRandomNumber generates a random 6-digit number as a string
func generateRandomNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // Generate a 6-digit random number between 000000 and 999999
}
