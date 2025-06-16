package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"envyous_token_backend/pkg/models"
)

// TaskHandler содержит зависимости, например подключение к БД, для handler-функций
type TaskHandler struct {
	DB *sql.DB
}

// GetDailyTasks - обработчик GET /tasks/daily?chat_id=...
// Возвращает список ежедневных задач (макс. 8 задач при первом запросе, затем по 4)
// с полем done=true/false для каждой задачи (выполнена сегодня или нет).
func (h *TaskHandler) GetDailyTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем параметр chat_id из URL
	chatIDParam := r.URL.Query().Get("chat_id")
	if chatIDParam == "" {
		http.Error(w, `{"error": "chat_id is required"}`, http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseInt(chatIDParam, 10, 64)
	if err != nil {
		http.Error(w, `{"error": "invalid chat_id"}`, http.StatusBadRequest)
		return
	}

	// Реализуем логику "8 заданий, затем по 4".
	// Можно использовать query-параметр ?offset для постраничной загрузки:
	limit := 8
	offset := 0
	offParam := r.URL.Query().Get("offset")
	if offParam != "" {
		offVal, err := strconv.Atoi(offParam)
		if err == nil && offVal >= 0 {
			offset = offVal
			if offset > 0 {
				limit = 4 // после первой страницы возвращаем по 4 задания
			}
		}
	}

	// Запрашиваем задания из БД (с флагом выполнено/не выполнено)
	tasks, err := models.GetDailyTasks(h.DB, chatID, limit, offset)
	if err != nil {
		log.Println("DB error in GetDailyTasks:", err)
		http.Error(w, `{"error": "failed to fetch tasks"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем список задач в формате JSON
	// (в виде массива объектов TaskWithStatus)
	json.NewEncoder(w).Encode(tasks)
}

// CompleteTask - обработчик POST /tasks/complete
// Помечает задание как выполненное пользователем и начисляет награду.
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ожидаем JSON с полями chat_id и task_id в теле запроса
	var reqBody struct {
		ChatID int64 `json:"chat_id"`
		TaskID int   `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	chatID := reqBody.ChatID
	taskID := reqBody.TaskID
	if chatID == 0 || taskID == 0 {
		http.Error(w, `{"error": "chat_id and task_id are required"}`, http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли такое задание и узнаем награду за него
	reward, err := models.GetTaskReward(h.DB, taskID)
	if err != nil {
		http.Error(w, `{"error": "task not found"}`, http.StatusNotFound)
		return
	}

	// Проверяем, не выполнено ли уже это задание сегодня
	done, err := models.HasCompletedToday(h.DB, chatID, taskID)
	if err != nil {
		log.Println("DB error in HasCompletedToday:", err)
		http.Error(w, `{"error": "db error"}`, http.StatusInternalServerError)
		return
	}
	if done {
		// Задание уже выполнено сегодня – предотвращаем повторное начисление
		http.Error(w, `{"error": "task already completed today"}`, http.StatusConflict)
		return
	}

	// Начинаем транзакцию, чтобы выполнить несколько шагов атомарно
	tx, err := h.DB.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		http.Error(w, `{"error": "server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // откат транзакции при ошибке (если не будет Commit)

	// 1. Добавляем запись о выполнении задания
	err = models.AddTaskCompletion(tx, chatID, taskID)
	if err != nil {
		// Если нарушение уникальности (запись уже есть), вернем конфликт
		log.Println("AddTaskCompletion error:", err)
		http.Error(w, `{"error": "could not record completion (maybe already done)"}`, http.StatusConflict)
		return
	}

	// 2. Начисляем награду: обновляем баланс пользователя и пишем транзакцию
	// Обновляем баланс пользователя в таблице users и получаем новый баланс
	var newBalance int
	err = tx.QueryRow(`UPDATE users SET balance = balance + $1 WHERE chat_id = $2 RETURNING balance;`,
		reward, chatID).Scan(&newBalance)
	if err != nil {
		log.Println("Failed to update user balance:", err)
		http.Error(w, `{"error": "failed to update user balance"}`, http.StatusInternalServerError)
		return
	}
	// Добавляем запись в таблицу transactions (логирование транзакции награды)
	_, err = tx.Exec(`INSERT INTO transactions (chat_id, amount, description) VALUES ($1, $2, $3);`,
		chatID, reward, "Daily task reward")
	if err != nil {
		log.Println("Failed to insert transaction record:", err)
		http.Error(w, `{"error": "failed to log transaction"}`, http.StatusInternalServerError)
		return
	}

	// 3. Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		log.Println("Transaction commit failed:", err)
		http.Error(w, `{"error": "server error"}`, http.StatusInternalServerError)
		return
	}

	// Формируем ответ: подтверждаем успех и возвращаем информацию о награде
	response := struct {
		Message    string `json:"message"`
		TaskID     int    `json:"task_id"`
		Reward     int    `json:"reward"`
		NewBalance int    `json:"new_balance"`
	}{
		Message:    "Task completed",
		TaskID:     taskID,
		Reward:     reward,
		NewBalance: newBalance,
	}
	json.NewEncoder(w).Encode(response)
}

// ClaimFarmingReward - обработчик POST /farming
// Начисляет пользователю 180 монет "за фарм" не чаще, чем раз в 3 минуты.
func (h *TaskHandler) ClaimFarmingReward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ожидаем JSON с полем chat_id (идентификатор пользователя)
	var reqBody struct {
		ChatID int64 `json:"chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	chatID := reqBody.ChatID
	if chatID == 0 {
		http.Error(w, `{"error": "chat_id is required"}`, http.StatusBadRequest)
		return
	}

	// Проверяем, когда в последний раз пользователь получал награду фарминга
	lastTime, exists, err := models.GetLastFarmed(h.DB, chatID)
	if err != nil {
		log.Println("DB error in GetLastFarmed:", err)
		http.Error(w, `{"error": "db error"}`, http.StatusInternalServerError)
		return
	}
	if exists {
		// Если запись есть, проверим разницу во времени
		elapsed := time.Since(lastTime)
		if elapsed < 3*time.Minute {
			// Прошло меньше 3 минут с последнего получения награды
			http.Error(w, `{"error": "farming reward already claimed, try later"}`, http.StatusTooManyRequests)
			return
		}
	}
	// Если записи нет или прошло >=3 минут, можно выдать награду.

	// Начинаем транзакцию для атомарного обновления времени и начисления монет
	tx, err := h.DB.Begin()
	if err != nil {
		log.Println("Failed to begin transaction (farming):", err)
		http.Error(w, `{"error": "server error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Обновляем/вставляем время последнего фарминга = сейчас
	err = models.UpdateLastFarmed(tx, chatID)
	if err != nil {
		log.Println("UpdateLastFarmed error:", err)
		http.Error(w, `{"error": "failed to update last_farmed time"}`, http.StatusInternalServerError)
		return
	}

	// 2. Начисляем 180 монет пользователю (обновляем баланс и логируем транзакцию)
	const rewardCoins = 180
	var newBalance int
	err = tx.QueryRow(`UPDATE users SET balance = balance + $1 WHERE chat_id = $2 RETURNING balance;`,
		rewardCoins, chatID).Scan(&newBalance)
	if err != nil {
		log.Println("Failed to update user balance (farming):", err)
		http.Error(w, `{"error": "failed to update user balance"}`, http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(`INSERT INTO transactions (chat_id, amount, description) VALUES ($1, $2, $3);`,
		chatID, rewardCoins, "Farming reward")
	if err != nil {
		log.Println("Failed to insert transaction (farming):", err)
		http.Error(w, `{"error": "failed to log transaction"}`, http.StatusInternalServerError)
		return
	}

	// 3. Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		log.Println("Farming transaction commit failed:", err)
		http.Error(w, `{"error": "server error"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем результат в JSON: 180 монет выдано, новый баланс
	response := struct {
		Message    string `json:"message"`
		Reward     int    `json:"reward"`
		NewBalance int    `json:"new_balance"`
	}{
		Message:    "Farming reward claimed",
		Reward:     rewardCoins,
		NewBalance: newBalance,
	}
	json.NewEncoder(w).Encode(response)
}
