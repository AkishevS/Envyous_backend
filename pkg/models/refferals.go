package models

//
//import (
//	"envyous_token_backend/middleware"
//	"envyous_token_backend/pkg/db"
//	"net/http"
//)
//
//func GetReferrals(w http.ResponseWriter, r *http.Request) {
//	userID := r.Context().Value(middleware.UserIDKey).(int64)
//
//	rows, err := db.Query(`
//        SELECT u.first_name, u.points, u.registered_at
//        FROM users u
//        WHERE u.invited_by = (SELECT id FROM users WHERE telegram_id = $1)`, userID)
//	if err != nil { /* ... */
//	}
//
//}
