package myapp

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"myapp/config"
	"net/http"
	"time"
)

// ฟังก์ชันสำหรับออกรางวัล
func DrawPrizesAll(w http.ResponseWriter, r *http.Request) {
	// ตรวจสอบว่ามีการออกรางวัลอยู่แล้วหรือไม่
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM winning_numbers").Scan(&count)
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการตรวจสอบการออกรางวัล:", err)
		http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
		return
	}

	// ถ้ามีการออกรางวัลแล้ว จะไม่สามารถออกรางวัลใหม่ได้
	if count > 0 {
		http.Error(w, "ไม่สามารถออกรางวัลได้เนื่องจากมีการออกรางวัลแล้ว", http.StatusBadRequest)
		return
	}

	// ดึงเลขหวยที่ถูกซื้อไปแล้ว
	rows, err := config.DB.Query("SELECT lid, lotto_number FROM lottery")
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลเลขหวยที่ถูกซื้อ:", err)
		http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// เก็บข้อมูลเลขหวยที่ถูกซื้อไว้ในแผนที่ (map)
	purchasedNumbers := make(map[string]int)
	for rows.Next() {
		var lottoNumber string
		var lid int
		if err := rows.Scan(&lid, &lottoNumber); err != nil {
			log.Println("เกิดข้อผิดพลาดในการอ่านข้อมูลเลขหวย:", err)
			http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
			return
		}
		purchasedNumbers[lottoNumber] = lid
	}

	// ดึงเลขหวยที่ถูกรางวัลก่อนหน้านี้
	existingWinners := make(map[string]bool)
	rows, err = config.DB.Query("SELECT lotto_number FROM winning_numbers")
	if err != nil {
		log.Println("เกิดข้อผิดพลาดในการดึงข้อมูลเลขหวยที่ถูกรางวัลก่อนหน้า:", err)
		http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var lottoNumber string
		if err := rows.Scan(&lottoNumber); err != nil {
			log.Println("เกิดข้อผิดพลาดในการอ่านข้อมูลเลขหวยที่ถูกรางวัลก่อนหน้า:", err)
			http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
			return
		}
		existingWinners[lottoNumber] = true
	}

	// สุ่มเลขรางวัลที่ 1-5 จากเลขที่ถูกซื้อไปแล้วและยังไม่เคยถูกรางวัล
	prizes := []WinningNumber{}
	prizeAmounts := []int{1000000, 500000, 100000, 50000, 10000}

	for _, prizeAmount := range prizeAmounts {
		prize, err := getUniqueRandomPrize1(purchasedNumbers, existingWinners, prizeAmount)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการสุ่มรางวัล:", err)
			http.Error(w, "ข้อผิดพลาดในการออกรางวัล", http.StatusInternalServerError)
			return
		}
		prizes = append(prizes, prize)
		existingWinners[prize.LottoNumber] = true // เพิ่มเลขนี้เข้าไปในรายการเลขที่ถูกรางวัลแล้ว
	}

	// เริ่มต้น transaction
	tx, err := config.DB.Begin()
	if err != nil {
		log.Println("เริ่มต้น transaction ไม่สำเร็จ:", err)
		http.Error(w, "ข้อผิดพลาดภายในระบบ", http.StatusInternalServerError)
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

	// บันทึกเลขรางวัลที่ 1-5 ลงในฐานข้อมูล
	for _, prize := range prizes {
		_, err = tx.Exec("INSERT INTO winning_numbers (lotto_number, prize_amount, lid) VALUES (?, ?, ?)",
			prize.LottoNumber, prize.PrizeAmount, prize.Lid)
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการบันทึกข้อมูลรางวัล:", err)
			return
		}
	}

	// ส่งผลลัพธ์กลับไปยังผู้ใช้
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "ออกรางวัลสำเร็จ",
		"prizes":  prizes,
	})
}

// ฟังก์ชันสำหรับสุ่มเลขรางวัลที่ไม่เคยถูกรางวัลมาก่อน
func getUniqueRandomPrize1(numbers map[string]int, existingWinners map[string]bool, prizeAmount int) (WinningNumber, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	keys := make([]string, 0, len(numbers))
	for key := range numbers {
		if !existingWinners[key] { // เลขนี้ไม่เคยถูกรางวัล
			keys = append(keys, key)
		}
	}

	if len(keys) == 0 {
		return WinningNumber{}, fmt.Errorf("ไม่มีเลขที่สามารถออกรางวัลได้")
	}

	// ใช้ตัวแปร r แทน rand สำหรับการสุ่ม
	selectedNumber := keys[r.Intn(len(keys))]
	return WinningNumber{
		LottoNumber: selectedNumber,
		PrizeAmount: prizeAmount,
		Lid:         numbers[selectedNumber],
	}, nil
}
