package handlers

//
//import (
//	"encoding/json"
//	"net/http"
//	"strings"
//
//	"envyous_token_backend/pkg/db"
//	"envyous_token_backend/pkg/models"
//	"golang.org/x/crypto/bcrypt"
//)
//
//func RegisterHandler(w http.ResponseWriter, r *http.Request) {
//	var reqUser models.User
//	if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
//		http.Error(w, "Invalid JSON data in request", http.StatusBadRequest)
//		return
//	}
//
//	reqUser.Username = strings.TrimSpace(reqUser.Username)
//	reqUser.Password = strings.TrimSpace(reqUser.Password)
//
//	if reqUser.ChatID == 0 || reqUser.Username == "" || reqUser.Password == "" {
//		http.Error(w, "Missing required fields", http.StatusBadRequest)
//		return
//	}
//
//	var exists bool
//	err := db.DB.QueryRow(
//		"SELECT EXISTS(SELECT 1 FROM users WHERE chat_id=$1 OR username=$2)",
//		reqUser.ChatID, reqUser.Username).Scan(&exists)
//	if err != nil {
//		http.Error(w, "Database error", http.StatusInternalServerError)
//		return
//	}
//	if exists {
//		http.Error(w, "User already exists", http.StatusConflict)
//		return
//	}
//
//	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqUser.Password), bcrypt.DefaultCost)
//	if err != nil {
//		http.Error(w, "Error hashing password", http.StatusInternalServerError)
//		return
//	}
//	reqUser.Password = string(hashedPassword)
//	reqUser.Balance = 0
//
//	var newUserID int64
//	query := "INSERT INTO users (chat_id, username, password, balance) VALUES ($1, $2, $3, $4) RETURNING id"
//	err = db.DB.QueryRow(query, reqUser.ChatID, reqUser.Username, reqUser.Password, reqUser.Balance).Scan(&newUserID)
//	if err != nil {
//		http.Error(w, "Insert error", http.StatusInternalServerError)
//		return
//	}
//	reqUser.ID = newUserID
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(http.StatusCreated)
//	json.NewEncoder(w).Encode(reqUser)
//}
