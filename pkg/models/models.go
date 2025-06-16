package models

import (
    "database/sql"
    "time"
)

type User struct {
    ID        int64        `json:"id"`
    ChatID    int64        `json:"chat_id"`
    Username  string       `json:"username"`
    Password  string       `json:"-"`
    Balance   int64        `json:"balance"`
    LastBonus sql.NullTime `json:"-"`
}

type Transaction struct {
    ID           int64     `json:"id"`
    FromUserID   int64     `json:"from_user_id"`
    ToUserID     int64     `json:"to_user_id"`
    Amount       int64     `json:"amount"`
    Type         string    `json:"type"`
    Timestamp    time.Time `json:"timestamp"`
    FromUsername string    `json:"from_username,omitempty"`
    ToUsername   string    `json:"to_username,omitempty"`
}
