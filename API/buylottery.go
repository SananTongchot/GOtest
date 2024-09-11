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
		UserID    int `json:"uid"` // รหัสผู้ใช้ที่ซื้อหวย
		LotteryID int `json:"lid"` // รหัสของหวยที่ถูกซื้อ
	}

	if err := json.NewDecoder(r.Body).Decode(&purchaseRequest); err != nil {
		log.Println("ข้อมูลที่ได้รับไม่ถูกต้อง:", err)
		http.Error(w, "ข้อมูลไม่ถูกต้อง", http.StatusBadRequest)
		return
	}

	// ดึงข้อมูลหวย
	var lottery model.Lottery
	err := config.DB.QueryRow("SELECT lid, lotto_number, sold, price FROM lottery WHERE lid = ?", purchaseRequest.LotteryID).Scan(
		&lottery.LID, &lottery.LottoNumber, &lottery.Sold, &lottery.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ไม่พบหวยที่มีรหัส: %d\n", purchaseRequest.LotteryID)
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

	// ดึงข้อมูลเครดิตของผู้ใช้
	var userCredit int
	err = config.DB.QueryRow("SELECT credit FROM user WHERE uid = ?", purchaseRequest.UserID).Scan(&userCredit)
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

	// ตรวจสอบเครดิตผู้ใช้ว่ามีเพียงพอหรือไม่
	if userCredit < lottery.Price {
		log.Printf("เครดิตไม่เพียงพอสำหรับผู้ใช้: %d, จำนวนที่ต้องการ: %d, เครดิตที่มี: %d\n", purchaseRequest.UserID, lottery.Price, userCredit)
		http.Error(w, "เครดิตไม่เพียงพอ", http.StatusBadRequest)
		return
	}

	// สร้าง Transaction ใหม่
	newTransaction := model.Transaction{
		UserID:        purchaseRequest.UserID,
		AmountPrice:   lottery.Price,
		AmountLottery: 1,
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

	// อัปเดตเครดิตของผู้ใช้
	newCredit := userCredit - lottery.Price
	_, err = tx.Exec("UPDATE user SET credit = ? WHERE uid = ?", newCredit, purchaseRequest.UserID)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการอัปเดตเครดิตของผู้ใช้:", err)
		return
	}

	// อัปเดตสถานะหวยเป็นขายแล้ว
	_, err = tx.Exec("UPDATE lottery SET sold = 1 WHERE lid = ?", purchaseRequest.LotteryID)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการอัปเดตสถานะหวย:", err)
		return
	}

// บันทึกข้อมูลลงในตาราง transactions
_, err = tx.Exec("INSERT INTO transactions (uid, lid, amount_price, amount_lottery) VALUES (?, ?, ?, ?)",
    newTransaction.UserID, purchaseRequest.LotteryID, newTransaction.AmountPrice, newTransaction.AmountLottery)
if err != nil {
    log.Println("เกิดข้อผิดพลาดในการบันทึกข้อมูลการทำรายการ:", err)
    return
}


	// ส่งผลลัพธ์กลับไปยังผู้ใช้
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "ซื้อหวยสำเร็จ",
		"newCredit": newCredit,
	})
}
