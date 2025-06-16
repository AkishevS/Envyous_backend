package models

import (
	"database/sql"
	"time"
)

// TaskCompletion представляет запись в таблице task_completions
type TaskCompletion struct {
	ID          int64     `json:"id"`
	ChatID      int64     `json:"chat_id"`
	TaskID      int       `json:"task_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// HasCompletedToday проверяет, выполнял ли пользователь (chatID) данное задание (taskID) сегодня.
func HasCompletedToday(db *sql.DB, chatID int64, taskID int) (bool, error) {
	var count int
	// Считаем количество записей выполнения за сегодня
	query := `
        SELECT COUNT(*) FROM task_completions 
        WHERE chat_id = $1 AND task_id = $2
          AND completed_at::date = current_date;
    `
	err := db.QueryRow(query, chatID, taskID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddTaskCompletion добавляет запись о выполнении задания (chatID, taskID) с текущим временем.
func AddTaskCompletion(tx *sql.Tx, chatID int64, taskID int) error {
	// Выполняем INSERT внутри транзакции tx
	_, err := tx.Exec(`
        INSERT INTO task_completions (chat_id, task_id, completed_at) 
        VALUES ($1, $2, NOW());
    `, chatID, taskID)
	return err
}
