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

func InitDB() {
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

func CloseDB() {

	if db != nil {
		err := db.Close()
		if err != nil {
			log.Println("Ошибка при закрытии подключения к БД:", err)
		} else {
			log.Println("Подключение к БД закрыто:", err)
		}
	}
}

func AddDataToDB(records []models.Product) (*models.Result, error) {

	if records == nil {
		log.Println("Переданный аргумент функции AddDataToDB равен nil.")
		return nil, nil
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Ошибка создании транзакции:", err)
		return nil, err
	}

	for _, val := range records {

		_, err = tx.Exec(`INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)`, val.Name, val.Category, val.Price, val.CreationTime)
		if err != nil {
			log.Println("Ошибка при внесении кортежа в БД:", err)
			tx.Rollback()
			return nil, err
		}
	}

	row := tx.QueryRow("SELECT COUNT(category) AS totalCategories, SUM(price) AS totalPrice FROM prices")

	result := models.Result{}
	result.TotalItems = len(records)

	if err := row.Scan(&result.TotalCategories, &result.TotalPrice); err != nil {

		log.Println("Ошибка при получении данных из кортежа:", err)
		tx.Rollback()

		return nil, err
	}

	err = tx.Commit()
	return &result, err

}

func GetDataFromDB() (string, error) {

	rows, err := db.Query("SELECT id, name, category, price, create_date FROM prices")
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Таблица price пуста.")
		} else {
			log.Println("Ошибка при закрытии rows:", err)

		}
		return "", err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			log.Println("Ошибка при завершении цикла обхода полученных из БД данных:", err)
		}
	}()

	var productSlice []models.Product

	for rows.Next() {
		p := models.Product{}
		if err := rows.Scan(&p.Id, &p.Name, &p.Category, &p.Price, &p.CreationTime); err != nil {
			log.Println("Ошибка сканирования строки:", err)
		} else {
			productSlice = append(productSlice, p)
		}
	}

	if err = rows.Err(); err != nil {
		log.Println("Цикл обхода данных из БД завершился с ошибкой:", err)
		return "", err
	}

	zipFilePath, err := helpers.SliceToZip(productSlice)

	return zipFilePath, err

}
