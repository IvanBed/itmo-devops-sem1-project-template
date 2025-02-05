package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"project_sem/internal/db"
	"project_sem/internal/helpers"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)

	case http.MethodPost:
		handlePost(w, r)
	}

}
func handleGet(w http.ResponseWriter, r *http.Request) {

	filePath, err := db.GetDataFromDB()
	if err != nil {
		http.Error(w, "Ошибка в получении данных и БД", http.StatusInternalServerError)
		log.Println("Ошибка в получении данных и БД:", err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=arch.zip")

	http.ServeFile(w, r, filePath)

	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Ошибка при удалении файла", http.StatusInternalServerError)
		log.Println("Ошибка при удалении файла:", err)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при чтении файла из запроса", http.StatusBadRequest)
		log.Println("Ошибка при чтении файла из запроса:", err)
		return
	}
	defer file.Close()

	log.Printf("Получен файл: %s", handler.Filename)

	dataSlice, err := helpers.ProcessZIPData(file)

	if err != nil {
		http.Error(w, "Ошибка при чтении ZIP архива в слайс.", http.StatusInternalServerError)
		log.Println("Ошибка при чтении ZIP архива в слайс:", err)
		return
	}

	resultJSON, err := db.AddDataToDB(dataSlice)
	if err != nil {
		http.Error(w, "Ошибка при добавлении данных в БД", http.StatusInternalServerError)
		log.Println("Ошибка при добавлении данных в БД:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resultJSON); err != nil {
		http.Error(w, "Ошибка в создании JSON файла", http.StatusInternalServerError)
		log.Println("Ошибка в создании JSON файла", err)
		return
	}

}
