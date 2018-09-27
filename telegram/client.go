package telegram

import (
	"encoding/json"
	"log"
	"net/http"

	"gopkg.in/resty.v1"
)

type TelegramClient struct {
	appKey string
}

func NewTelegramClient(appKey string) (*TelegramClient, error) {
	return &TelegramClient{appKey: appKey}, nil
}

func (client TelegramClient) SendMessage(chatId interface{}, text string) error {
	_, err := resty.R().
		SetBody(struct {
			chatId interface{} `json:"chat_id"`
			text   string      `json:"text"`
		}{chatId: chatId, text: text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}

func (server TelegramClient) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var update Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Println(err)
	}
	err = server.SendMessage(update.Message.Chat.ID, update.Message.Text)
	if err != nil {
		log.Println(err)
	}
}

func (server TelegramClient) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	w.Header()["Content-type"] = []string{"text/plain"}
}
