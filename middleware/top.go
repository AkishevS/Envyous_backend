package middleware

//
//import (
//	"envyous_token_backend/pkg/db"
//	"net/http"
//)
//
//func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
//	_, err := db.Query(`
//        SELECT first_name, points
//        FROM users
//        ORDER BY points DESC
//        LIMIT 10`)
//	if err != nil { /* ... */
//	}
//
//	// вернуть JSON с участниками
//}
