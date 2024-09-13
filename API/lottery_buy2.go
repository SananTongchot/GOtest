package myapp

import (
	"encoding/json"
	"log"
	"myapp/config"
	"net/http"
)

type PurchasedLotteryWithWin struct {
	Lid         int    `json:"lid"`
	LottoNumber string `json:"lotto_number"`
	Win         bool   `json:"win"`
	PrizeAmount int    `json:"prize_amount"`
	PrizeRank   string `json:"prize_rank"`
}

func GetPurchasedLotteriesByUID2(w http.ResponseWriter, r *http.Request) {
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
		SELECT 
			t.lid, 
			l.lotto_number, 
			l.win, 
			COALESCE(w.prize_amount, 0) AS prize_amount, 
			CASE 
				WHEN w.prize_amount = 1000000 THEN 'รางวัลที่ 1'
				WHEN w.prize_amount = 500000 THEN 'รางวัลที่ 2'
				WHEN w.prize_amount = 100000 THEN 'รางวัลที่ 3'
				WHEN w.prize_amount = 50000 THEN 'รางวัลที่ 4'
				WHEN w.prize_amount = 10000 THEN 'รางวัลที่ 5'
				ELSE 'ไม่ได้รับรางวัล'
			END AS prize_rank
		FROM 
			transactions t
		INNER JOIN 
			lottery l ON t.lid = l.lid
		LEFT JOIN 
			winning_numbers w ON l.lotto_number = w.lotto_number
		WHERE 
			t.uid = ?;
	`

	rows, err := config.DB.Query(query, reqBody.UID)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลลอตเตอรี่:", err)
		http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// เก็บข้อมูลลอตเตอรี่ที่ดึงมาใน slice และใช้ map เพื่อตรวจสอบข้อมูลซ้ำ
	var lotteries []PurchasedLotteryWithWin
	seen := make(map[int]bool) // map เพื่อตรวจสอบว่า lid ถูกดึงมาแล้วหรือยัง

	for rows.Next() {
		var lottery PurchasedLotteryWithWin
		err := rows.Scan(&lottery.Lid, &lottery.LottoNumber, &lottery.Win, &lottery.PrizeAmount, &lottery.PrizeRank)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการอ่านข้อมูลลอตเตอรี่:", err)
			http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
			return
		}

		// ตรวจสอบว่ามี lid นี้ใน map แล้วหรือไม่ ถ้ายังไม่มีให้เพิ่มเข้าไป
		if !seen[lottery.Lid] {
			seen[lottery.Lid] = true
			lotteries = append(lotteries, lottery)
		}
	}

	// ตรวจสอบว่ามีลอตเตอรี่ที่ผู้ใช้ซื้อหรือไม่
	if len(lotteries) == 0 {
		http.Error(w, "ไม่พบลอตเตอรี่ที่ผู้ใช้ซื้อไป", http.StatusNotFound)
		return
	}

	// ส่งข้อมูลลอตเตอรี่ที่ถูกซื้อไปแล้วกลับไป
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lotteries)
}
