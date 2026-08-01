package main

import (
	"fmt"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Guest"
	}
	response := fmt.Sprintf("Hello, %s!", name)
	w.Write([]byte(response))
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
