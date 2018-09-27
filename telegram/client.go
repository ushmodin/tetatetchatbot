package telegram

import (
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
			ChatId interface{} `json:"chat_id"`
			Text   string      `json:"text"`
		}{ChatId: chatId, Text: text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}
