package myapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"myapp/config"
	"myapp/model"
	"net/http"
)

func BuyLottery(w http.ResponseWriter, r *http.Request) {
	var purchaseRequest struct {
		UserID     int   `json:"uid"`  // รหัสผู้ใช้ที่ซื้อหวย
		LotteryIDs []int `json:"lids"` // รายการรหัสหวยที่ต้องการซื้อ
	}

	if err := json.NewDecoder(r.Body).Decode(&purchaseRequest); err != nil {
		log.Println("ข้อมูลที่ได้รับไม่ถูกต้อง:", err)
		http.Error(w, "ข้อมูลไม่ถูกต้อง", http.StatusBadRequest)
		return
	}
	// ตรวจสอบว่ามีรายการหวยที่ต้องการซื้อหรือไม่
	if len(purchaseRequest.LotteryIDs) == 0 {
		log.Println("ไม่มีรายการหวยที่ต้องการซื้อ")
		http.Error(w, "ไม่มีรายการหวยที่ต้องการซื้อ", http.StatusBadRequest)
		return
	}

	// ดึงข้อมูลเครดิตของผู้ใช้
	var userCredit int
	err := config.DB.QueryRow("SELECT credit FROM user WHERE uid = ?", purchaseRequest.UserID).Scan(&userCredit)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ไม่พบผู้ใช้ที่มีรหัส: %d\n", purchaseRequest.UserID)
			http.Error(w, "ไม่พบผู้ใช้", http.StatusNotFound)
		} else {
			log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลเครดิตของผู้ใช้:", err)
			http.Error(w, "ข้อผิดพลาดภายในระบบ", http.StatusInternalServerError)
		}
		return
	}

	// เริ่มต้น transaction
	tx, err := config.DB.Begin()
	if err != nil {
		log.Println("เริ่มต้น transaction ไม่สำเร็จ:", err)
		http.Error(w, "เริ่มต้น transaction ไม่สำเร็จ", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			tx.Commit()
		}
	}()

	totalCost := 0
	for _, lotteryID := range purchaseRequest.LotteryIDs {
		// ดึงข้อมูลหวยแต่ละรายการ
		var lottery model.Lottery
		err := tx.QueryRow("SELECT lid, lotto_number, sold, price FROM lottery WHERE lid = ?", lotteryID).Scan(
			&lottery.LID, &lottery.LottoNumber, &lottery.Sold, &lottery.Price)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("ไม่พบหวยที่มีรหัส: %d\n", lotteryID)
				http.Error(w, "ไม่พบหวย", http.StatusNotFound)
			} else {
				log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลหวย:", err)
				http.Error(w, "ข้อผิดพลาดภายในระบบ", http.StatusInternalServerError)
			}
			return
		}

		// ตรวจสอบว่าหวยขายไปแล้วหรือไม่
		if lottery.Sold {
			log.Printf("หวยถูกขายไปแล้ว: %d\n", lottery.LID)
			http.Error(w, "หวยถูกขายไปแล้ว", http.StatusBadRequest)
			return
		}

		// เพิ่มราคาหวยลงในราคาทั้งหมด
		totalCost += lottery.Price

		// อัปเดตสถานะหวยเป็นขายแล้ว
		_, err = tx.Exec("UPDATE lottery SET sold = 1 WHERE lid = ?", lotteryID)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการอัปเดตสถานะหวย:", err)
			return
		}

		// บันทึกข้อมูลลงในตาราง transactions พร้อมกับ lotto_number
		_, err = tx.Exec("INSERT INTO transactions (uid, lid, lotto_number, amount_price, amount_lottery) VALUES (?, ?, ?, ?, ?)",
			purchaseRequest.UserID, lotteryID, lottery.LottoNumber, lottery.Price, 1)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการบันทึกข้อมูลการทำรายการ:", err)
			return
		}
	}

	// ตรวจสอบว่าเครดิตของผู้ใช้เพียงพอหรือไม่
	if userCredit < totalCost {
		log.Printf("เครดิตไม่เพียงพอสำหรับผู้ใช้: %d, จำนวนที่ต้องการ: %d, เครดิตที่มี: %d\n", purchaseRequest.UserID, totalCost, userCredit)
		http.Error(w, "เครดิตไม่เพียงพอ", http.StatusBadRequest)
		return
	}

	// อัปเดตเครดิตของผู้ใช้
	newCredit := userCredit - totalCost
	_, err = tx.Exec("UPDATE user SET credit = ? WHERE uid = ?", newCredit, purchaseRequest.UserID)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการอัปเดตเครดิตของผู้ใช้:", err)
		return
	}

	// ส่งผลลัพธ์กลับไปยังผู้ใช้
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "ซื้อหวยสำเร็จ",
		"newCredit": newCredit,
	})
}
