package main

import (
	"log"
	"net/http"
	"project_sem/internal/db"
	"project_sem/internal/handlers"
)

func main() {

	db.InitDB()
	http.HandleFunc("/api/v0/prices", handlers.Handler)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Println("Ошибка при запуске сервера:", err)
	}
	db.CloseDB()
}
