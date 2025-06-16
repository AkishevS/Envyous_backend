package handlers

//
//import (
//	"db/sql"
//	"encoding/json"
//	"net/http"
//)
//
//func TelegramAuthHandler(w http.ResponseWriter, r *http.Request) {
//	var req struct {
//		ChatID   int64  `json:"chat_id"`
//		Username string `json:"username"`
//	}
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//		http.Error(w, "Invalid JSON", http.StatusBadRequest)
//		return
//	}
//
//	var user User
//	err := DB.QueryRow(`SELECT id, chat_id, username, balance FROM users WHERE chat_id=$1`, req.ChatID).
//		Scan(&user.ID, &user.ChatID, &user.Username, &user.Balance)
//
//	if err == sql.ErrNoRows {
//		err = DB.QueryRow(`
//			INSERT INTO users (chat_id, username, balance)
//			VALUES ($1, $2, 0)
//			RETURNING id, chat_id, username, balance
//		`, req.ChatID, req.Username).Scan(&user.ID, &user.ChatID, &user.Username, &user.Balance)
//		if err != nil {
//			http.Error(w, "Error creating user", http.StatusInternalServerError)
//			return
//		}
//	} else if err != nil {
//		http.Error(w, "DB error", http.StatusInternalServerError)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(user)
//}
