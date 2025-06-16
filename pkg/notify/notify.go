package notify

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func SendMessage(chatID int64, message string) {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	jsonBody := []byte(fmt.Sprintf(`{"chat_id":%d,"text":"%s"}`, chatID, message))

	http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
}
