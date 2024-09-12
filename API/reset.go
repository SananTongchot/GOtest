package myapp

import (
	"database/sql"
	"log"
	"net/http"
)

// ResetHandler handles the reset operation
func ResetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ตรวจสอบว่าเป็นคำขอ POST
		if r.Method != http.MethodPost {
			http.Error(w, "กรุณาใช้ method POST", http.StatusMethodNotAllowed)
			return
		}

		// เริ่มต้นการลบข้อมูล

		// 1. ลบข้อมูลล็อตเตอรี่ทั้งหมดที่สามารถซื้อได้
		_, err := db.Exec("DELETE FROM lottery")
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลบล็อตเตอรี่:", err)
			http.Error(w, "ไม่สามารถลบล็อตเตอรี่ได้", http.StatusInternalServerError)
			return
		}

		// 2. ลบข้อมูลผลการจับรางวัลทั้งหมด
		_, err = db.Exec("DELETE FROM winning_numbers")
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลบผลการจับรางวัล:", err)
			http.Error(w, "ไม่สามารถลบผลการจับรางวัลได้", http.StatusInternalServerError)
			return
		}

		// 3. ลบบัญชีผู้ใช้ทั้งหมด ยกเว้นแอดมิน
		_, err = db.Exec("DELETE FROM user WHERE type = 2")
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลบบัญชีผู้ใช้:", err)
			http.Error(w, "ไม่สามารถลบบัญชีผู้ใช้ได้", http.StatusInternalServerError)
			return
		}

		// 4. ลบข้อมูลการซื้อทั้งหมด
		_, err = db.Exec("DELETE FROM transactions")
		if err != nil {
			log.Println("เกิดข้อผิดพลาดในการลบข้อมูลการซื้อ:", err)
			http.Error(w, "ไม่สามารถลบข้อมูลการซื้อได้", http.StatusInternalServerError)
			return
		}

		// ส่งข้อความยืนยันการรีเซ็ตเสร็จสิ้น
		w.Write([]byte("รีเซ็ตข้อมูลทั้งหมดเสร็จสิ้น"))

		// เรียกใช้ GenerateLotteryHandler หลังจากการรีเซ็ตเสร็จสิ้น
		GenerateLotteryHandler(db)(w, r)
	}
}
