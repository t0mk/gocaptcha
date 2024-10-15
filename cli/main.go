package main

import (
	"fmt"
	"net/http"

	"github.com/t0mk/gocaptcha"
)

func main() {
	fmt.Println("Starting server on :8080")
	fmt.Println("Access the captcha at: http://localhost:8080/getcaptcha")
	fmt.Println("Verify captcha at: http://localhost:8080/verify?code=YOUR_CODE")

	// Set up routes
	http.HandleFunc("/getcaptcha", gocaptcha.CaptchaHandler)
	http.HandleFunc("/verify", gocaptcha.CaptchaHandler)

	// Start the HTTP server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
