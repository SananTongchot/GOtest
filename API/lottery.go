package myapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"myapp/model"
	"net/http"
	"time"
)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
// GenerateLotteryHandler returns an HTTP handler function that generates lottery numbers
func GenerateLotteryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		const numNumbers = 100
		lottoNumbers := make([]string, 0, numNumbers)
		uniqueNumbers := make(map[string]bool)

		// ดึงเลขหวยที่มีอยู่แล้วจากฐานข้อมูลเพื่อตรวจสอบว่าไม่ซ้ำ
		existingNumbers := make(map[string]bool)
		rows, err := db.Query("SELECT lotto_number FROM lottery")
		if err != nil {
			http.Error(w, "Failed to query existing numbers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var number string
			if err := rows.Scan(&number); err != nil {
				http.Error(w, "Failed to read existing numbers", http.StatusInternalServerError)
				return
			}
			existingNumbers[number] = true
		}

		// Generate unique lottery numbers
		for len(lottoNumbers) < numNumbers {
			lottoNumber := generateRandomNumber()
			if _, exists := uniqueNumbers[lottoNumber]; !exists && !existingNumbers[lottoNumber] {
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
				log.Println("Failed to insert lottery number:", err)
				return // return immediately to trigger rollback
			}
		}
	}
}

// generateRandomNumber generates a random 6-digit number as a string
func generateRandomNumber() string {
	return fmt.Sprintf("%06d", rng.Intn(1000000)) // สร้างหมายเลขสุ่ม 6 หลักระหว่าง 000000 และ 999999
}

