package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"project_sem/internal/helpers"
	"project_sem/internal/models"
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
}

func AddDataToDB(CSVFilePath string) (*models.Result, error) {
	fmt.Println("test0")
	records, err := helpers.CSVToSlice(CSVFilePath)
	if err != nil {
		return nil, err
	}
	fmt.Println("test1")
	tx, err := db.Begin()
	if err != nil {
		
		return nil, err
	}
	fmt.Println("test11")
	for _, val := range records {

		_, err = tx.Exec(`INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)`, val.Name, val.Category, val.Price, val.CreationTime)
		if err != nil {
			tx.Rollback()
			fmt.Println("test2")
			return nil, err
		}
	}
	fmt.Println("test3")
	row := tx.QueryRow("SELECT COUNT(category) AS totalCategories, SUM(price) AS totalPrice FROM prices")

	if err != nil {

	}
	fmt.Println("test4")
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("")
		} else {
			tx.Rollback()
			return nil, err
		}
	}

	result := models.Result{}
	result.TotalItems = len(records)
	if err := row.Scan(&result.TotalCategories, &result.TotalPrice); err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	return &result, err

}

func GetDataFromDB() (string, error) {

	rows, err := db.Query("SELECT id, name, category, price, create_date FROM prices")
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer rows.Close()

	var productSlice []models.Product

	for rows.Next() {
		p := models.Product{}
		if err := rows.Scan(&p.Id, &p.Name, &p.Category, &p.Price, &p.CreationTime); err != nil {
			continue
		} else {
			productSlice = append(productSlice, p)
		}
	}

	//Добавить обработку ошибок
	if err = rows.Err(); err != nil {
		log.Println(err)
	}

	if err = rows.Close(); err != nil {
		log.Println(err)
	}

	zipFilePath, err := helpers.SliceToZip(productSlice)
	return zipFilePath, err

}
