package myapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	
)

// โครงสร้างข้อมูลสำหรับรับ JSON input

// ตรวจสอบว่าหมายเลขหวยที่ผู้ใช้ซื้อถูกรางวัลหรือไม่
func CheckUserLotteryResultsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// ตรวจสอบ method ของ request ให้แน่ใจว่าเป็น POST
		if r.Method != http.MethodPost {
			http.Error(w, "กรุณาใช้ method POST", http.StatusMethodNotAllowed)
			return
		}

		// แปลง JSON input ไปเป็นโครงสร้างข้อมูล UserRequest
		var userRequest UserRequest
		if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
			http.Error(w, "ไม่สามารถอ่านข้อมูล JSON ได้", http.StatusBadRequest)
			return
		}

		// ตรวจสอบว่า uid ไม่ว่าง
		if userRequest.UID == 0 {
			http.Error(w, "กรุณาระบุ UID ของผู้ใช้", http.StatusBadRequest)
			return
		}

		// ดึงหมายเลขลอตเตอรี่ทั้งหมดที่ผู้ใช้ซื้อจากตาราง transactions และ lottery พร้อมกับ lid
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

		// เก็บหมายเลขลอตเตอรี่ที่ผู้ใช้ซื้อไว้ใน slice
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

		// ตรวจสอบหมายเลขทั้งหมดที่ผู้ใช้ซื้อว่าถูกรางวัลหรือไม่
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
			} else if err != sql.ErrNoRows {
				log.Println("เกิดข้อผิดพลาดในการตรวจสอบรางวัล:", err)
				http.Error(w, "ข้อผิดพลาดในการตรวจสอบหมายเลขลอตเตอรี่", http.StatusInternalServerError)
				return
			}
		}

		// ส่งผลลัพธ์กลับไปยังผู้ใช้
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "การตรวจสอบรางวัลเสร็จสิ้น",
			"winning_numbers": winningNumbers,
		})
	}
}
