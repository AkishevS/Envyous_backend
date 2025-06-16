package models

import (
	"database/sql"
	"time"
)

// FarmingReward представляет запись в таблице farming_rewards
type FarmingReward struct {
	ChatID     int64     `json:"chat_id"`
	LastFarmed time.Time `json:"last_farmed"`
}

// GetLastFarmed возвращает время последнего фарминга для пользователя,
// а также признак наличия записи.
func GetLastFarmed(db *sql.DB, chatID int64) (time.Time, bool, error) {
	var last time.Time
	err := db.QueryRow("SELECT last_farmed FROM farming_rewards WHERE chat_id = $1", chatID).Scan(&last)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, false, nil // записи нет
		}
		return time.Time{}, false, err // ошибка запроса
	}
	return last, true, nil
}

// UpdateLastFarmed обновляет время последнего фарминга на текущее (NOW()).
// Если запись для пользователя отсутствует, создаёт новую.
func UpdateLastFarmed(tx *sql.Tx, chatID int64) error {
	// Используем INSERT ... ON CONFLICT для upsert: обновить или вставить запись
	_, err := tx.Exec(`
        INSERT INTO farming_rewards(chat_id, last_farmed) 
        VALUES ($1, NOW()) 
        ON CONFLICT (chat_id) DO UPDATE SET last_farmed = EXCLUDED.last_farmed;
    `, chatID)
	return err
}
