package myapp

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"myapp/config"
	"myapp/model"
)

// RequestBody โครงสร้างข้อมูลสำหรับรับจาก POST
type RequestBody struct {
	UID int `json:"uid"`
}

// GetUserByUIDPost รับ UID จาก POST body และดึงข้อมูลผู้ใช้
func GetaUser(w http.ResponseWriter, r *http.Request) {
	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.UID == 0 {
		http.Error(w, "รูปแบบข้อมูลไม่ถูกต้อง หรือไม่ได้ระบุ UID", http.StatusBadRequest)
		return
	}

	var user model.User
	err = config.DB.QueryRow("SELECT uid, username, phone, email, password, credit, type FROM user WHERE uid = ?", reqBody.UID).Scan(
		&user.UID, &user.Username, &user.Phone, &user.Email, &user.Password, &user.Credit, &user.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "ไม่พบผู้ใช้", http.StatusNotFound)
		} else {
			log.Println("เกิดข้อผิดพลาดในการค้นหาข้อมูลผู้ใช้:", err)
			http.Error(w, "ข้อผิดพลาดภายในเซิร์ฟเวอร์", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
