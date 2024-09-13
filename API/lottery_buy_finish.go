package myapp

import (
	"encoding/json"
	"log"
	"myapp/config"
	"net/http"
)

// โครงสร้างข้อมูลสำหรับการรับค่า UID จาก POST
type RequestUID struct {
	UID int `json:"uid"`
}

// โครงสร้างข้อมูลสำหรับลอตเตอรี่ที่ถูกซื้อไปแล้ว
type PurchasedLottery struct {
	Lid           int    `json:"lid"`
	LottoNumber   string `json:"lotto_number"`
	AmountPrice   int    `json:"amount_price"`
	AmountLottery int    `json:"amount_lottery"`
	Win           bool   `json:"win"` // เพิ่มคอลัมน์ win
}

// ฟังก์ชันสำหรับแสดงลอตเตอรี่ทั้งหมดที่ผู้ใช้ซื้อไปแล้ว (POST)
func GetPurchasedLotteriesByUID(w http.ResponseWriter, r *http.Request) {
	// ตรวจสอบว่าเป็น POST request หรือไม่
	if r.Method != http.MethodPost {
		http.Error(w, "กรุณาใช้ POST request", http.StatusBadRequest)
		return
	}

	// ดึงข้อมูล UID จาก POST body
	var reqBody RequestUID
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.UID == 0 {
		http.Error(w, "ข้อมูลไม่ถูกต้อง หรือไม่ได้ระบุ UID", http.StatusBadRequest)
		return
	}

	// Query ข้อมูลลอตเตอรี่ที่ผู้ใช้ซื้อไปแล้ว
	query := `
		SELECT t.lid, t.lotto_number, t.amount_price, t.amount_lottery, l.win
		FROM transactions t
		INNER JOIN lottery l ON t.lid = l.lid
		WHERE t.uid = ?;
	`

	rows, err := config.DB.Query(query, reqBody.UID)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลลอตเตอรี่:", err)
		http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// เก็บข้อมูลลอตเตอรี่ที่ดึงมาใน slice
	var lotteries []PurchasedLottery
	for rows.Next() {
		var lottery PurchasedLottery
		err := rows.Scan(&lottery.Lid, &lottery.LottoNumber, &lottery.AmountPrice, &lottery.AmountLottery, &lottery.Win) // เพิ่มการสแกนคอลัมน์ win
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการอ่านข้อมูลลอตเตอรี่:", err)
			http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
			return
		}
		lotteries = append(lotteries, lottery)
	}

	// ตรวจสอบว่ามีลอตเตอรี่ที่ถูกซื้อหรือไม่
	if len(lotteries) == 0 {
		http.Error(w, "ไม่พบลอตเตอรี่ที่ผู้ใช้ซื้อไป", http.StatusNotFound)
		return
	}

	// ส่งข้อมูลลอตเตอรี่ที่ถูกซื้อไปแล้วกลับไป
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lotteries)
}
