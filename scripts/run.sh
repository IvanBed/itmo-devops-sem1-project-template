#!/bin/bash

APP_NAME="go-api-server"

DB_NAME="project_sem_1"   
DB_USERNAME="validator"  
DB_PASSWORD="val1dat0r"  
DB_PORT=5432              

APP_PORT=8080             

echo "Запуск приложения \"$APP_NAME\" на локальном сервере..."
echo "Параметры:"
echo " - База данных: $DB_NAME"
echo " - Пользователь: $DB_USERNAME"
echo " - Порт базы данных: $DB_PORT"
echo " - Порт приложения: $APP_PORT"

 
if ! command -v go &> /dev/null; then
    echo "Ошибка: Go не установлен! Убедитесь, что Go установлен и добавлен в PATH."
    exit 1
fi

 
echo "Обновление зависимостей..."
go mod tidy

 
echo "Сборка приложения..."
go build -o "$APP_NAME" main.go

 
if [ $? -ne 0 ]; then
    echo "Ошибка: Сборка приложения \"$APP_NAME\" завершилась с ошибкой."
    exit 1
fi
 

echo "Запускаю приложение \"$APP_NAME\" на порту $APP_PORT..."
./"$APP_NAME" &
APP_PID=$! 

 
if ps -p $APP_PID > /dev/null; then
    echo "Приложение \"$APP_NAME\" успешно запущено на порту: $APP_PORT"
else
    echo "Ошибка запуска приложения \"$APP_NAME\"."
    exit 1
fi
