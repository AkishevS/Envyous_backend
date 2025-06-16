package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"envyous_token_backend/pkg/db"
	"envyous_token_backend/pkg/models"
	"github.com/gorilla/mux"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatIDStr := vars["chat_id"]
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	var user models.User
	err = db.DB.QueryRow(`SELECT id, chat_id, username, balance FROM users WHERE chat_id = $1`, chatID).
		Scan(&user.ID, &user.ChatID, &user.Username, &user.Balance)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := db.DB.Query(`
        SELECT t.id, t.from_user_id, t.to_user_id, t.amount, t.type, t.timestamp,
               f.username, u.username
        FROM transactions t
        LEFT JOIN users f ON t.from_user_id = f.id
        LEFT JOIN users u ON t.to_user_id = u.id
        WHERE t.from_user_id = $1 OR t.to_user_id = $1
        ORDER BY t.timestamp DESC LIMIT 3`, user.ID)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tr models.Transaction
		var fromName, toName sql.NullString
		rows.Scan(&tr.ID, &tr.FromUserID, &tr.ToUserID, &tr.Amount, &tr.Type, &tr.Timestamp, &fromName, &toName)
		if fromName.Valid {
			tr.FromUsername = fromName.String
		} else if tr.Type == "bonus" {
			tr.FromUsername = "System"
		}
		if toName.Valid {
			tr.ToUsername = toName.String
		}
		transactions = append(transactions, tr)
	}

	// Возвращаем JSON
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chat_id":     user.ChatID,
		"username":    user.Username,
		"balance":     user.Balance,
		"last_events": transactions,
	})
}
