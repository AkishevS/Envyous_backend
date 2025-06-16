package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"envyous_token_backend/middleware"
)

type Handlers struct {
	DB *sql.DB
}

func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{DB: db}
}

func (h *Handlers) Leaderboard(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.ContextUserIDKey)
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.QueryContext(r.Context(), `
        SELECT inviter_telegram_id, COUNT(*) AS invites_count
        FROM referrals
        GROUP BY inviter_telegram_id
        ORDER BY invites_count DESC
        LIMIT 10`)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type LeaderItem struct {
		TelegramID   int64 `json:"telegram_id"`
		InvitesCount int   `json:"invites_count"`
	}
	var leaderboard []LeaderItem
	for rows.Next() {
		var item LeaderItem
		if err := rows.Scan(&item.TelegramID, &item.InvitesCount); err != nil {
			continue
		}
		leaderboard = append(leaderboard, item)
	}
	json.NewEncoder(w).Encode(leaderboard)
}

func (h *Handlers) Referrals(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.ContextUserIDKey)
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	telegramID := userIDVal.(int64)

	rows, err := h.DB.QueryContext(r.Context(),
		`SELECT invitee_telegram_id FROM referrals WHERE inviter_telegram_id = $1`, telegramID)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var referrals []int64
	for rows.Next() {
		var inviteeID int64
		if err := rows.Scan(&inviteeID); err != nil {
			continue
		}
		referrals = append(referrals, inviteeID)
	}
	json.NewEncoder(w).Encode(referrals)
}
