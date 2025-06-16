package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Не удалось загрузить файл .env (может быть, приложение запущено в продакшене):", err)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	if host == "" || user == "" || dbname == "" {
		log.Fatal("Database configuration not set in .env")
	}

	// Формирование строки подключения
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		log.Fatal("БД недоступна:", err)
	}

	log.Println("Успешно подключено к базе данных")
	DB = db
	return DB
}
