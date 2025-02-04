package helpers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"project_sem/internal/models"
	"strconv"
	"strings"
)

func CSVToSlice(CSVFilePath string) ([]models.Product, error) {

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

	records := make([]models.Product, 0)

	if _, err := reader.Read(); err != nil {
		log.Println("Ошибка при чтении файла:", err)
		return nil, err
	}

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
			log.Println("Ошибка при парсинге id")
			id = 0
		}

		name := tokens[1]
		category := tokens[2]
		price, err := strconv.ParseFloat(tokens[3], 64)
		creationTime := tokens[4]
		if err != nil {
			log.Println("Ошибка при парсинге price")
			price = 0
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

	return records, nil
}

func SliceToZip(rows []models.Product) (string, error) {

	CSVdatafile, err := os.Create("data.csv")
	if err != nil {
		log.Println("Ошбика в создании CSV файла.")
		return "", err
	}
	defer func() {
		if err := CSVdatafile.Close(); err != nil {
			log.Println("Ошибка при попытке закрыть файл", err)
		}
	}()
	defer func() {
		if err := os.Remove(CSVdatafile.Name()); err != nil {
			log.Println("Ошибка при удалении файла", err)
		}
	}()

	writer := csv.NewWriter(CSVdatafile)
	defer writer.Flush()

	headers := []string{"id", "name", "category", "price", "create_date"}
	if err := writer.Write(headers); err != nil {
		log.Println("Ошибка при записи заголовка в CSV файл")
		return "", nil
	}

	for _, row := range rows {
		strRow := []string{fmt.Sprintf("%d", row.Id), row.Name, row.Category, fmt.Sprintf("%.2f", row.Price), row.CreationTime}
		if err := writer.Write(strRow); err != nil {
			log.Println("Ошибка при запиcи строки в CSV файл")
			continue
		}

	}
	writer.Flush()
	zipFile, err := os.Create("data.zip")
	if err != nil {
		log.Println("Ошибка при создании архива zip")
		return "", nil
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			log.Println("Ошибка при попытке закрыть файл архива", err)
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Дописать вывод в лог ошибок

	fileInfo, err := CSVdatafile.Stat()
	if err != nil {
		log.Println("Ошибка при получении информации о файле:", err)
		return "", nil
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.Println("Ошибка при создании заголовка zip файла:", err)
		return "", nil
	}
	header.Name = CSVdatafile.Name()

	writerPart, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Println("Ошибка при создании части записи для zip архива:", err)
		return "", nil
	}

	CSVdatafile.Seek(0, 0)
	_, err = io.Copy(writerPart, CSVdatafile)
	if err != nil {
		log.Println("Ошибка при копировании данных в часть zip архива:", err)
		return "", nil
	}

	log.Println("Добавлен в архив:", CSVdatafile.Name())

	return zipFile.Name(), nil
}
