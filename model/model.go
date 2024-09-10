package model

type User struct {
	UID      int    `json:"uid"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"` // "1" for admin, "2" for user
	Credit   int    `json:"credit"`
}

type Lottery struct {
	LID         int    `json:"lid"`          // คอลัมน์ lid, กำหนด auto increment ในฐานข้อมูล
	LottoNumber string `json:"lotto_number"` // เลขลอตเตอรีในรูปแบบ string
	Sold        bool   `json:"sold"`         // สถานะการขาย
	Price       int    `json:"price"`        // ราคาของลอตเตอรี
}

type LotteryResponse struct {
	Message string `json:"message"` // ข้อความแสดงผล
	Number  string `json:"number"`  // เลขลอตเตอรีที่สุ่มได้
}
