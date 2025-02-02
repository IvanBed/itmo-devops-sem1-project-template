package db

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"project_sem/internal/models"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	HOST := os.Getenv("POSTGRES_HOST")
	PORT := os.Getenv("POSTGRES_PORT")
	DATABASE := os.Getenv("POSTGRES_DB")
	USER := os.Getenv("POSTGRES_USER")
	PASSWORD := os.Getenv("POSTGRES_PASSWORD")

	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", HOST, PORT, USER, PASSWORD, DATABASE)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Ошибка при подключении к БД:", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
}

func AddDataToDB(CSVFilePath string) (*models.Result, error) {

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

	totalItems := 0
	totalCategories := 0
	totalPrice := 0.0

	categoriesMap := make(map[string]int)
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
		fmt.Println(tokens[0])
		_, err = db.Exec(`INSERT INTO prices (id, name, category, price, create_date) VALUES ($1, $2, $3, $4, $5)`, tokens[0], tokens[1], tokens[2], tokens[3], tokens[4])
		if err != nil {
			log.Println("Ошибка при внесении данных в БД:", err)
			return nil, err
		}
		categoriesMap[tokens[2]] = 1

		price, _ := strconv.ParseFloat(tokens[3], 64)

		totalItems++
		totalPrice = totalPrice + price
	}
	totalCategories = len(categoriesMap)
	result := models.Result{
		TotalItems:      totalItems,
		TotalCategories: totalCategories,
		TotalPrice:      totalPrice,
	}

	return &result, nil
}

func GetDataFromDB() (string, error) {

	CSVdatafile, err := os.Create("data.csv")
	if err != nil {
		return "", nil
	}
	defer CSVdatafile.Close()

	rows, err := db.Query("SELECT id, name, category, price, create_date FROM prices")
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer rows.Close()

	writer := csv.NewWriter(CSVdatafile)
	defer writer.Flush()

	headers := []string{"id", "name", "category", "price", "create_date"}
	if err := writer.Write(headers); err != nil {
		return "", nil
	}

	for rows.Next() {
		p := models.Product{}
		if err := rows.Scan(&p.Id, &p.Name, &p.Category, &p.Price, &p.CreationTime); err != nil {

			continue
		}

		strRow := []string{fmt.Sprintf("%d", p.Id), p.Name, p.Category, fmt.Sprintf("%.2f", p.Price), p.CreationTime}
		if err := writer.Write(strRow); err != nil {

			continue
		}
	}

	writer.Flush()

	zipFile, err := os.Create("data.zip")
	if err != nil {
		return "", nil
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileInfo, err := CSVdatafile.Stat()
	if err != nil {
		return "", nil
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return "", nil
	}
	header.Name = CSVdatafile.Name()

	writerPart, err := zipWriter.CreateHeader(header)
	if err != nil {
		return "", nil
	}

	CSVdatafile.Seek(0, 0)
	_, err = io.Copy(writerPart, CSVdatafile)
	if err != nil {
		return "", nil
	}

	fmt.Println("Добавлен в архив:", CSVdatafile.Name())
	fmt.Println("тест:", zipFile.Name())
	os.Remove(CSVdatafile.Name())
	return zipFile.Name(), nil
}
