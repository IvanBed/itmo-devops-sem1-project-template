#!/bin/bash

DB_NAME="project-sem-1"
DB_USER="validator"
DB_PASSWORD="val1dat0r"
DB_TABLE="prices"
PG_HOST="localhost"  
PG_PORT="5432"

function setup_database() {
    echo "Настройка базы данных..."

    PGPASSWORD=$DB_PASSWORD psql -U $DB_USER -h $PG_HOST -p $PG_PORT -d $DB_NAME <<EOF
    CREATE TABLE IF NOT EXISTS prices
    (
        id SERIAL PRIMARY KEY,
        create_date timestamp without time zone,
        name character varying(1000) COLLATE pg_catalog."default",
        category character varying(1000) COLLATE pg_catalog."default",
        price numeric
    );

    ALTER TABLE IF EXISTS prices
    OWNER TO ${DB_USER};

    GRANT DELETE, UPDATE, INSERT, SELECT ON TABLE prices TO ${DB_USER};
EOF

    if [ $? -eq 0 ]; then
        echo "Таблица '${DB_TABLE}' успешно создана или уже существует."
        echo "Пользователь '${DB_USER}' имеет права на запись в таблицу."
    else
        echo "Произошла ошибка при создании таблицы '${DB_TABLE}'."
        exit 1
    fi
}

function install_dependencies() {
    echo "Установка зависимостей приложения..."
    if [ -f "go.mod" ]; then
        go mod tidy
        echo "Зависимости установлены."
    else
        echo "Файл go.mod не найден. Пропускаем установку зависимостей."
    fi
}

echo "Запуск скрипта подготовки..."

setup_database
install_dependencies

echo "Скрипт подготовки завершён успешно."
