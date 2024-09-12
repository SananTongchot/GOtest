package myapp

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"myapp/model"
)

// GetUnpurchasedLotteriesHandler คืนค่า HTTP handler function ที่ใช้ในการดึงข้อมูลลอตเตอรี่ที่ยังไม่ถูกซื้อ
func GetUnpurchasedLotteriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// ดึงข้อมูลลอตเตอรี่ที่ยังไม่ถูกซื้อ (sold = 0)
		rows, err := db.Query("SELECT lid, lotto_number, price FROM lottery WHERE sold = 0")
		if err != nil {
			http.Error(w, "เกิดข้อผิดพลาดในการดึงข้อมูลลอตเตอรี่ที่ยังไม่ถูกซื้อ", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// เก็บข้อมูลลอตเตอรี่ที่ยังไม่ถูกซื้อไว้ใน slice
		var lotteries []model.Lottery
		for rows.Next() {
			var lottery model.Lottery
			if err := rows.Scan(&lottery.LID, &lottery.LottoNumber, &lottery.Price); err != nil {
				http.Error(w, "เกิดข้อผิดพลาดในการอ่านข้อมูลลอตเตอรี่", http.StatusInternalServerError)
				return
			}
			lotteries = append(lotteries, lottery)
		}

		// ตรวจสอบข้อผิดพลาดเพิ่มเติมในการอ่านข้อมูล
		if err := rows.Err(); err != nil {
			http.Error(w, "เกิดข้อผิดพลาดในการอ่านข้อมูลลอตเตอรี่", http.StatusInternalServerError)
			return
		}

		// ส่งข้อมูลลอตเตอรี่ที่ยังไม่ถูกซื้อกลับไปยังผู้ใช้
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(lotteries)
	}
}
