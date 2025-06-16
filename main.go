package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"envyous_token_backend/handlers"
	"envyous_token_backend/routes"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	log.Println("Подключение к базе данных успешно")

	r := mux.NewRouter()
	r.Use(withCORS)

	taskHandler := &handlers.TaskHandler{DB: db}
	routes.RegisterTaskRoutes(r, taskHandler)

	log.Println("Запуск сервера на порту 8080...")
	http.ListenAndServe(":8080", r)
}
