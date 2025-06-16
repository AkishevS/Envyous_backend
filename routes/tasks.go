package routes

import (
	_ "net/http"

	"envyous_token_backend/handlers"
	"github.com/gorilla/mux"
)

func RegisterTaskRoutes(router *mux.Router, handler *handlers.TaskHandler) {

	router.HandleFunc("/tasks/daily", handler.GetDailyTasks).Methods("GET")
	router.HandleFunc("/tasks/complete", handler.CompleteTask).Methods("POST")
	router.HandleFunc("/farming", handler.ClaimFarmingReward).Methods("POST")
	router.HandleFunc("/tap", handler.TapCoin).Methods("POST")
}
