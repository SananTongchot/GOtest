// package main
// import (
//
//	"fmt"
//	"net/http"
//
// )
//
//	func main() {
//		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//			fmt.Fprint(w, "Hello World")
//		})
//			// รันเซิร์ฟเวอร์ที่พอร์ต 8080
//			http.ListenAndServe(":8080", nil)
//		}
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // ใช้ underscore (_) เพื่อ import driver โดยไม่ต้องเรียกใช้ตรงๆ
)

func main() {
	// ระบุข้อมูลการเชื่อมต่อ เช่น username, password, hostname, และ database name
	// รูปแบบ DSN: username:password@tcp(host:port)/dbname
	dsn := "web66_65011212243:65011212243@csmsu@tcp(202.28.34.197:3306)/web66_65011212243"

	// สร้าง connection pool ไปยัง MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		
		panic(err)
	}
	defer db.Close()

	// ตรวจสอบการเชื่อมต่อ
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("เชื่อมต่อฐานข้อมูลสำเร็จ!")
}
