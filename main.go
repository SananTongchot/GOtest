package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World")
	})

	// รันเซิร์ฟเวอร์ที่พอร์ต 8080
	http.ListenAndServe(":8080", nil)
}

// package main

// import (
// 	"fmt"
// 	"net/http"
// )

// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprint(w, "Hello world Lotto888")
// 	})
// 	fmt.Println("Server on port 5000")
// 	http.ListenAndServe(":5000", nil)
// }
