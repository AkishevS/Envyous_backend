package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// ContextKey – тип ключа для context.WithValue (для хранения данных пользователя).
type ContextKey string

// ContextUserIDKey – ключ в контексте для telegram_id пользователя.
const ContextUserIDKey ContextKey = "telegramUserID"

// ValidateInitData – middleware для проверки initData Telegram Mini App.
// db – подключение к БД; botToken – токен Telegram-бота.
func ValidateInitData(db *sql.DB, botToken string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовок Authorization: должен быть вид "tma {initDataRaw}"
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "tma ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			initDataRaw := strings.TrimPrefix(auth, "tma ")

			// Валидируем подпись initData по алгоритму Telegram (используем официальный пакет:contentReference[oaicite:3]{index=3})
			if err := initdata.Validate(initDataRaw, botToken, 24*time.Hour); err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Парсим initData в структуру, чтобы получить поля User и StartParam
			data, err := initdata.Parse(initDataRaw)
			if err != nil {
				http.Error(w, "Invalid initData", http.StatusBadRequest)
				return
			}
			// Telegram ID текущего пользователя
			telegramID := data.User.ID

			// Если передан start_param (id пригласившего), сохраняем связь в таблицу referrals
			if data.StartParam != "" {
				// Предполагается, что start_param – строковый telegram_id пригласившего
				if inviterID, err := strconv.ParseInt(data.StartParam, 10, 64); err == nil {
					if inviterID != telegramID {
						// Вставляем запись о приглашении (ON CONFLICT DO NOTHING, чтобы не дублировать)
						ctx := r.Context()
						_, _ = db.ExecContext(ctx,
							`INSERT INTO referrals (inviter_telegram_id, invitee_telegram_id)
                             VALUES ($1, $2) ON CONFLICT DO NOTHING`,
							inviterID, telegramID)
						// Ошибки логируем/игнорируем, чтобы не прерывать авторизацию
					}
				}
			}

			// Кладём telegram_id пользователя в контекст запроса для последующего использования в handlers
			ctx := context.WithValue(r.Context(), ContextUserIDKey, telegramID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
