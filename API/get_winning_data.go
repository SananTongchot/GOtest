package myapp

import (
	"encoding/json"
	"log"
	"net/http"

	"myapp/config"
)

// โครงสร้างข้อมูล WinningNumber
type WinningNumber2 struct {
	LottoNumber string `json:"lotto_number"`
	PrizeAmount int    `json:"prize_amount"`
	Lid         int    `json:"lid"`
}

// GetAllWinningNumbers ดึงข้อมูลทั้งหมดจากตาราง winning_numbers
func GetAllWinningNumbers(w http.ResponseWriter, r *http.Request) {
	// ดึงข้อมูลจากตาราง winning_numbers
	rows, err := config.DB.Query("SELECT lotto_number, prize_amount, lid FROM winning_numbers order by lid asc")
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลรางวัล:", err)
		http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// สร้าง slice สำหรับเก็บข้อมูลรางวัลทั้งหมด
	var winningNumbers []WinningNumber2

	// วนลูปอ่านข้อมูลแต่ละแถวและเก็บไว้ใน slice
	for rows.Next() {
		var wn WinningNumber2
		if err := rows.Scan(&wn.LottoNumber, &wn.PrizeAmount, &wn.Lid); err != nil {
			log.Println("เกิดข้อผิดพลาดในการอ่านข้อมูลรางวัล:", err)
			http.Error(w, "ข้อผิดพลาดในการดึงข้อมูล", http.StatusInternalServerError)
			return
		}
		winningNumbers = append(winningNumbers, wn)
	}

	// ตรวจสอบข้อผิดพลาดในการวนลูป (ถ้ามี)
	if err := rows.Err(); err != nil {
		log.Println("เกิดข้อผิดพลาดในการประมวลผลข้อมูล:", err)
		http.Error(w, "ข้อผิดพลาดในการดึงข้อมูล", http.StatusInternalServerError)
		return
	}

	// ส่งข้อมูลรางวัลทั้งหมดในรูปแบบ JSON กลับไปยังผู้ใช้
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(winningNumbers)
}
