package handlers

//
//import (
//   "db/sql"
//   "encoding/json"
//   "net/http"
//   "strconv"
//   "strings"
//   "time"
//
//   "github.com/gorilla/mux"
//   "envyous_token_backend/pkg/db"
//   "envyous_token_backend/pkg/models"
//)
//
//func BonusHandler(w http.ResponseWriter, r *http.Request) {
//   var req struct {
//       UserID int64 `json:"user_id"`
//   }
//   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//       http.Error(w, "Invalid JSON", http.StatusBadRequest)
//       return
//   }
//
//   bonusAmount := int64(100)
//   tx, err := db.DB.Begin()
//   if err != nil {
//       http.Error(w, "DB error", http.StatusInternalServerError)
//       return
//   }
//
//   var newBalance int64
//   err = tx.QueryRow(`
//       UPDATE users SET balance = balance + $2, last_bonus = NOW()
//       WHERE id = $1 AND (last_bonus IS NULL OR last_bonus <= NOW() - INTERVAL '8 HOURS')
//       RETURNING balance`, req.UserID, bonusAmount).Scan(&newBalance)
//
//   if err != nil {
//       tx.Rollback()
//       if err == sql.ErrNoRows {
//           http.Error(w, "Bonus not available yet", http.StatusBadRequest)
//       } else {
//           http.Error(w, "DB error", http.StatusInternalServerError)
//       }
//       return
//   }
//
//   _, err = tx.Exec(`
//       INSERT INTO transactions (from_user_id, to_user_id, amount, type, timestamp)
//       VALUES ($1, $2, $3, $4, $5)`, 0, req.UserID, bonusAmount, "bonus", time.Now())
//
//   if err != nil {
//       tx.Rollback()
//       http.Error(w, "Transaction error", http.StatusInternalServerError)
//       return
//   }
//
//   tx.Commit()
//   json.NewEncoder(w).Encode(map[string]interface{}{
//       "bonus_amount": bonusAmount,
//       "new_balance":  newBalance,
//   })
//}
//
//func TransferHandler(w http.ResponseWriter, r *http.Request) {
//   var req struct {
//       FromUserID int64  `json:"from_user_id"`
//       ToUsername string `json:"to_username"`
//       Amount     int64  `json:"amount"`
//   }
//   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//       http.Error(w, "Invalid JSON", http.StatusBadRequest)
//       return
//   }
//
//   var toUserID int64
//   err := db.DB.QueryRow("SELECT id FROM users WHERE username=$1", strings.TrimSpace(req.ToUsername)).Scan(&toUserID)
//   if err != nil {
//       http.Error(w, "Recipient not found", http.StatusNotFound)
//       return
//   }
//
//   if toUserID == req.FromUserID {
//       http.Error(w, "Cannot transfer to yourself", http.StatusBadRequest)
//       return
//   }
//
//   tx, err := db.DB.Begin()
//   if err != nil {
//       http.Error(w, "DB error", http.StatusInternalServerError)
//       return
//   }
//
//   var newBalance int64
//   err = tx.QueryRow(
//       "UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1 RETURNING balance",
//       req.Amount, req.FromUserID).Scan(&newBalance)
//   if err != nil {
//       tx.Rollback()
//       http.Error(w, "Insufficient balance or sender error", http.StatusBadRequest)
//       return
//   }
//
//   _, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", req.Amount, toUserID)
//   if err != nil {
//       tx.Rollback()
//       http.Error(w, "Recipient update failed", http.StatusInternalServerError)
//       return
//   }
//
//   _, err = tx.Exec("INSERT INTO transactions (from_user_id, to_user_id, amount, type, timestamp) VALUES ($1, $2, $3, $4, $5)",
//       req.FromUserID, toUserID, req.Amount, "transfer", time.Now())
//   if err != nil {
//       tx.Rollback()
//       http.Error(w, "Transaction record failed", http.StatusInternalServerError)
//       return
//   }
//
//   tx.Commit()
//   json.NewEncoder(w).Encode(map[string]interface{}{
//       "transferred":  req.Amount,
//       "new_balance":  newBalance,
//       "to_user_id":   toUserID,
//       "from_user_id": req.FromUserID,
//   })
//}
//
//func HistoryHandler(w http.ResponseWriter, r *http.Request) {
//   vars := mux.Vars(r)
//   idStr := vars["user_id"]
//   userID, err := strconv.ParseInt(idStr, 10, 64)
//   if err != nil {
//       http.Error(w, "Invalid user_id", http.StatusBadRequest)
//       return
//   }
//
//   rows, err := db.DB.Query(`
//       SELECT t.id, t.from_user_id, t.to_user_id, t.amount, t.type, t.timestamp,
//              f.username, u.username
//       FROM transactions t
//       LEFT JOIN users f ON t.from_user_id = f.id
//       LEFT JOIN users u ON t.to_user_id = u.id
//       WHERE t.from_user_id = $1 OR t.to_user_id = $1
//       ORDER BY t.timestamp DESC`, userID)
//   if err != nil {
//       http.Error(w, "DB error", http.StatusInternalServerError)
//       return
//   }
//   defer rows.Close()
//
//   var history []models.Transaction
//   for rows.Next() {
//       var tr models.Transaction
//       var fromName, toName sql.NullString
//       rows.Scan(&tr.ID, &tr.FromUserID, &tr.ToUserID, &tr.Amount, &tr.Type, &tr.Timestamp, &fromName, &toName)
//       if fromName.Valid {
//           tr.FromUsername = fromName.String
//       } else if tr.Type == "bonus" {
//           tr.FromUsername = "System"
//       }
//       if toName.Valid {
//           tr.ToUsername = toName.String
//       }
//       history = append(history, tr)
//   }
//
//   json.NewEncoder(w).Encode(history)
//}
