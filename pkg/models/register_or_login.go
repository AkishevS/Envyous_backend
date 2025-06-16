package models

import "database/sql"

func RegisterOrLoginUser(db *sql.DB, userID int64, firstName, username, refCode string) (int, error) {
	var id int

	// если уже зарегистрирован
	err := db.QueryRow(`SELECT id FROM users WHERE telegram_id = $1`, userID).Scan(&id)
	if err == nil {
		return id, nil
	}

	tx, _ := db.Begin()

	// Найдём пригласившего
	var inviterID *int
	if refCode != "" {
		err = db.QueryRow(`SELECT id FROM users WHERE telegram_id = $1`, refCode).Scan(&id)
		if err == nil {
			inviterID = &id
		}
	}

	// создаём пользователя
	err = tx.QueryRow(`
        INSERT INTO users (telegram_id, first_name, username, invited_by)
        VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, firstName, username, inviterID).Scan(&id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// записываем в таблицу рефералов
	if inviterID != nil {
		_, err := tx.Exec(`INSERT INTO referrals (inviter_id, invitee_id) VALUES ($1, $2)`, *inviterID, id)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	return id, tx.Commit()
}
