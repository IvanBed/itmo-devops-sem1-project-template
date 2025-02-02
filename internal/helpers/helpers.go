package helpers

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"project_sem/internal/models"
	"strconv"
	"strings"
)

func ReadCSVToSlice(CSVFilePath string) ([]models.Product, error) {

	CSVFile, err := os.Open(CSVFilePath)
	if err != nil {
		log.Println("Ошибка при попытке открыть файл", err)
		return nil, err
	}

	defer func() {
		if err := CSVFile.Close(); err != nil {
			log.Println("Ошибка при попытке закрыть файл", err)
		}
	}()
	defer func() {
		if err := os.Remove(CSVFile.Name()); err != nil {
			log.Println("Ошибка при удалении файла:", err)
		}
	}()

	reader := csv.NewReader(CSVFile)
	reader.Comma = ';'

	if _, err := reader.Read(); err != nil {
		log.Println("Ошибка при чтении файла:", err)
		return nil, err
	}

	records := make([]models.Product)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Ошибка при чтении CSV файла:", err)
			return nil, err
		}
		tokens := strings.Split(record[0], ",")

		id, err := strconv.Atoi(tokens[0])
		if err != nil {
			log.Println("Ошибка при парсинге id", err)
			return nil, err
		}
		creationTime := tokens[1]
		name := tokens[2]
		category := tokens[3]
		price, err := strconv.ParseFloat(tokens[4], 64)
		if err != nil {
			log.Println("Ошибка при парсинге price", err)
			return nil, err
		}

		p := models.Product{
			Id:           id,
			CreationTime: creationTime,
			Name:         name,
			Category:     category,
			Price:        price,
		}
		records = append(records, p)

	}
	return records
}
