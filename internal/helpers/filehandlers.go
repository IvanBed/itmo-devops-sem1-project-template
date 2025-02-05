package helpers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"project_sem/internal/models"
	"strconv"
	"strings"
)

func CSVToSlice(CSVFile io.Reader) ([]models.Product, error) {

	if CSVFile == nil {
		log.Println("Файл не передан")
		return nil, nil
	}

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

func ProcessZIPData(r io.Reader) ([]models.Product, error) {

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, r); err != nil {
		log.Println("Ошибка при чтении файла", err)
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		log.Println("Ошибка при распаковке архива", err)
		return nil, err
	}

	var CSVFile io.ReadCloser
	for _, f := range zipReader.File {
		if filepath.Ext(f.Name) == ".csv" {

			CSVFile, err = f.Open()

			if err != nil {
				log.Println("Ошибка при открытии файла внутри архива", err)
				return nil, err
			}
			break
		}
	}

	if CSVFile == nil {
		log.Println("CSV файл не найден в архиве", err)
		return nil, err
	}

	dataSlice, err := CSVToSlice(CSVFile)
	if err != nil {
		log.Println("Ошибка при преобразовании CSV в слайс", err)
		return nil, err
	}

	CSVFile.Close()
	return dataSlice, nil
}
