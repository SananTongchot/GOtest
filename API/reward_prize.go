package myapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// Structure for handling JSON input
type UserRequest struct {
	UID int `json:"uid"`
}

// Handler to check lottery results and update user credits
func RewardPrize(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "กรุณาใช้ method POST", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON input into UserRequest
		var userRequest UserRequest
		if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
			http.Error(w, "ไม่สามารถอ่านข้อมูล JSON ได้", http.StatusBadRequest)
			return
		}

		// Ensure UID is provided
		if userRequest.UID == 0 {
			http.Error(w, "กรุณาระบุ UID ของผู้ใช้", http.StatusBadRequest)
			return
		}

		// Fetch all lottery numbers purchased by the user
		rows, err := db.Query(`
			SELECT l.lid, l.lotto_number 
			FROM transactions t
			JOIN lottery l ON t.lid = l.lid
			WHERE t.uid = ?`, userRequest.UID)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการดึงหมายเลขลอตเตอรี่ที่ผู้ใช้ซื้อ:", err)
			http.Error(w, "ข้อผิดพลาดในการตรวจสอบหมายเลขลอตเตอรี่", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Store purchased lottery numbers
		var userLottoNumbers []map[string]interface{}
		for rows.Next() {
			var lottoNumber string
			var lid int
			if err := rows.Scan(&lid, &lottoNumber); err != nil {
				log.Println("เกิดข้อผิดพลาดในการอ่านหมายเลขลอตเตอรี่:", err)
				http.Error(w, "ข้อผิดพลาดในการตรวจสอบหมายเลขลอตเตอรี่", http.StatusInternalServerError)
				return
			}
			userLottoNumbers = append(userLottoNumbers, map[string]interface{}{
				"lid":          lid,
				"lotto_number": lottoNumber,
			})
		}

		// Calculate total prize amount and store winning numbers
		var totalPrizeAmount int
		var winningNumbers []map[string]interface{}
		for _, lottoInfo := range userLottoNumbers {
			var prizeAmount int
			err := db.QueryRow(`
				SELECT wn.prize_amount 
				FROM winning_numbers wn 
				JOIN lottery l ON wn.lid = l.lid 
				WHERE l.lotto_number = ?`, lottoInfo["lotto_number"]).Scan(&prizeAmount)
			if err == nil {
				winningNumbers = append(winningNumbers, map[string]interface{}{
					"lid":          lottoInfo["lid"],
					"lotto_number": lottoInfo["lotto_number"],
					"prize_amount": prizeAmount,
				})
				totalPrizeAmount += prizeAmount

				// Update winning lottery status
				_, err = db.Exec(`
					UPDATE winning_numbers 
					SET status = 1 
					WHERE lotto_number = ?`, lottoInfo["lotto_number"])
				if err != nil {
					log.Println("เกิดข้อผิดพลาดในการอัปเดตสถานะหมายเลขลอตเตอรี่:", err)
					http.Error(w, "ข้อผิดพลาดในการอัปเดตสถานะหมายเลขลอตเตอรี่", http.StatusInternalServerError)
					return
				}
			} else if err != sql.ErrNoRows {
				log.Println("เกิดข้อผิดพลาดในการตรวจสอบรางวัล:", err)
				http.Error(w, "ข้อผิดพลาดในการตรวจสอบหมายเลขลอตเตอรี่", http.StatusInternalServerError)
				return
			}
		}

		// Fetch current credit before update
		var currentCredit int
		err = db.QueryRow(`
			SELECT credit 
			FROM user 
			WHERE uid = ?`, userRequest.UID).Scan(&currentCredit)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการดึงเครดิตปัจจุบันของผู้ใช้:", err)
			http.Error(w, "ข้อผิดพลาดในการดึงเครดิตปัจจุบัน", http.StatusInternalServerError)
			return
		}

		// Update user credit
		_, err = db.Exec(`
			UPDATE user
			SET credit = credit + ? 
			WHERE uid = ?`, totalPrizeAmount, userRequest.UID)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการเพิ่มเครดิตให้ผู้ใช้:", err)
			http.Error(w, "ข้อผิดพลาดในการเพิ่มเครดิต", http.StatusInternalServerError)
			return
		}

		// Fetch updated credit after the update
		var updatedCredit int
		err = db.QueryRow(`
			SELECT credit 
			FROM user 
			WHERE uid = ?`, userRequest.UID).Scan(&updatedCredit)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการดึงเครดิตล่าสุดของผู้ใช้:", err)
			http.Error(w, "ข้อผิดพลาดในการดึงเครดิตล่าสุด", http.StatusInternalServerError)
			return
		}

		// Send response to the user
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":            "การขึ้นเงินรางวัลเสร็จสิ้น",
			"winning_numbers":    winningNumbers,
			"previous_credit":    currentCredit,
			"total_prize_amount": totalPrizeAmount,
			"current_credit":     updatedCredit,
		})
	}
}
