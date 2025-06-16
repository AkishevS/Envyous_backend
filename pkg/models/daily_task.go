package models

import (
	"database/sql"
)

// DailyTask представляет запись задания из таблицы daily_tasks
type DailyTask struct {
	ID     int    `json:"id"`
	Emoji  string `json:"emoji"`
	Name   string `json:"name"`
	Reward int    `json:"reward"`
}

// TaskWithStatus объединяет информацию о задании и флаг выполнения
type TaskWithStatus struct {
	DailyTask      // встраиваем поля задачи
	Done      bool `json:"done"` // выполнено ли задание (сегодня)
}

// GetDailyTasks возвращает список ежедневных заданий (с ограничением limit/offset)
// и помечает их флагом Done, если пользователь (chatID) выполнил их сегодня.
func GetDailyTasks(db *sql.DB, chatID int64, limit, offset int) ([]TaskWithStatus, error) {
	// SQL-запрос выбирает задания и проверяет, выполнены ли они текущим пользователем сегодня
	query := `
        SELECT t.id, t.emoji, t.name, t.reward,
               CASE 
                 WHEN c.task_id IS NULL THEN FALSE 
                 ELSE TRUE 
               END as done
        FROM daily_tasks t
        LEFT JOIN task_completions c 
          ON t.id = c.task_id 
          AND c.chat_id = $1 
          AND c.completed_at::date = current_date
        ORDER BY t.id
        LIMIT $2 OFFSET $3;
    `
	rows, err := db.Query(query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []TaskWithStatus{}
	for rows.Next() {
		var task TaskWithStatus
		if err := rows.Scan(&task.ID, &task.Emoji, &task.Name, &task.Reward, &task.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// GetTaskReward получает величину награды для задания по его id.
// Возвращает reward и ошибку, если задание не найдено.
func GetTaskReward(db *sql.DB, taskID int) (int, error) {
	var reward int
	err := db.QueryRow("SELECT reward FROM daily_tasks WHERE id = $1", taskID).Scan(&reward)
	if err != nil {
		return 0, err // ошибка или sql.ErrNoRows, если не найдено
	}
	return reward, nil
}
