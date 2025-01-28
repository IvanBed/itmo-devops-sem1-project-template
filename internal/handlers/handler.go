package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"project_sem/internal/db"
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

	tempZipFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		http.Error(w, "Ошибка при создании временного файла", http.StatusInternalServerError)
		log.Println("Ошибка при создании временного файла:", err)
		return
	}
	defer os.Remove(tempZipFile.Name())
	defer tempZipFile.Close()

	log.Printf("Получен файл: %s", handler.Filename)

	if _, err := io.Copy(tempZipFile, file); err != nil {
		http.Error(w, "Ошибка при записи в временный файл", http.StatusInternalServerError)
		log.Println("Ошибка при записи в временный файл:", err)
		return
	}

	rZip, err := zip.OpenReader(tempZipFile.Name())
	if err != nil {
		http.Error(w, "Ошибка при открытии ZIP-файла", http.StatusInternalServerError)
		log.Println("Ошибка при открытии ZIP-файла:", err)
		return
	}
	defer rZip.Close()

	var csvFound bool
	var CSVFilePath string

	for _, f := range rZip.File {
		if filepath.Ext(f.Name) == ".csv" {
			outFile, err := os.Create("data.csv")
			if err != nil {
				http.Error(w, "Ошибка при создании CSV-файла", http.StatusInternalServerError)
				log.Println("Ошибка при создании CSV-файла:", err)
				return
			}
			defer outFile.Close()

			zipFile, err := f.Open()
			if err != nil {
				http.Error(w, "Ошибка при открытии файла внутри ZIP", http.StatusInternalServerError)
				log.Println("Ошибка при открытии файла внутри ZIP:", err)
				return
			}
			defer zipFile.Close()

			if _, err := io.Copy(outFile, zipFile); err != nil {
				http.Error(w, "Ошибка при копировании CSV-файла", http.StatusInternalServerError)
				log.Println("Ошибка при копировании CSV-файла:", err)
				return
			}
			csvFound = true
			CSVFilePath = outFile.Name()
			break
		}
	}

	if !csvFound {
		http.Error(w, "CSV-файл не найден в ZIP-архиве", http.StatusBadRequest)
		log.Println("CSV-файл не найден в ZIP-архиве")
		return
	}

	resultJSON, err := db.AddDataToDB(CSVFilePath)
	if err != nil {
		http.Error(w, "Ошибка при добавлении данных в БД", http.StatusInternalServerError)
		log.Println("Ошибка при добавлении данных в БД:", err)
		return
	}

	fmt.Println(resultJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resultJSON); err != nil {
		http.Error(w, "Ошибка в создании JSON файла", http.StatusInternalServerError)
		log.Println("Ошибка в создании JSON файла", err)
		return
	}

}

