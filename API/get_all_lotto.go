package myapp

import (
	"database/sql"
	"encoding/json"
	"myapp/model"
	"net/http"
)

// GetAllLotteriesHandler คืนค่า HTTP handler function ที่ใช้ในการดึงข้อมูลลอตเตอรี่ทั้งหมด
func GetAllLotteriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// ดึงข้อมูลลอตเตอรี่ทั้งหมด
		rows, err := db.Query("SELECT lid, lotto_number, price, sold FROM lottery ORDER BY lotto_number ASC")
		if err != nil {
			http.Error(w, "เกิดข้อผิดพลาดในการดึงข้อมูลลอตเตอรี่ทั้งหมด", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// เก็บข้อมูลลอตเตอรี่ทั้งหมดไว้ใน slice
		var lotteries []model.Lottery
		for rows.Next() {
			var lottery model.Lottery
			if err := rows.Scan(&lottery.LID, &lottery.LottoNumber, &lottery.Price, &lottery.Sold); err != nil {
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

		// ส่งข้อมูลลอตเตอรี่ทั้งหมดกลับไปยังผู้ใช้
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(lotteries)
	}
}
