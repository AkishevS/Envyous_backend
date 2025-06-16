package middleware

import "database/sql"

func AddPoints(db *sql.DB, userID int, earned int) error {
	_, err := db.Exec(`UPDATE users SET points = points + $1 WHERE id = $2`, earned, userID)
	if err != nil {
		return err
	}

	// добавим 10% рефереру
	var inviterID int
	err = db.QueryRow(`SELECT invited_by FROM users WHERE id = $1`, userID).Scan(&inviterID)
	if err == nil && inviterID != 0 {
		bonus := earned / 10
		_, _ = db.Exec(`UPDATE users SET points = points + $1 WHERE id = $2`, bonus, inviterID)
	}
	return nil
}
