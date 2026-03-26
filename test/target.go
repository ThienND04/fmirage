package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	cnt := 0
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 1. Giả lập các thư mục nhạy cảm (Trả về 200 OK)
		if path == "/admin" || path == "/config.php" || path == "/api" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("BINGO! Ban da tim thay thu muc an!"))
			return
		}

		// 2. Giả lập một trang bị chặn bởi WAF/CAPTCHA (Luôn trả về 200 OK, nhưng size cố định)
		// Giả sử size của trang báo lỗi này luôn là 100 bytes
		if strings.HasPrefix(path, "/fake_") {
			w.WriteHeader(http.StatusOK)
			fakeHTML := fmt.Sprintf("%-100s", "<html><body>Please verify you are human (Fake CAPTCHA)</body></html>")
			w.Write([]byte(fakeHTML))
			return
		}

		// 3. Các đường dẫn còn lại trả về 404
		cnt++
		fmt.Printf("[%d] Received request for: %s\n", cnt, path)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	fmt.Println("[+] Mục tiêu  đang chạy tại: http://localhost:8080")

	// Khởi chạy server trên cổng 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Lỗi chạy server: %v\n", err)
	}
}
