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

func (client TelegramClient) SendServiceMessage(chatId interface{}, text string) error {
	_, err := resty.R().
		SetBody(struct {
			ChatId interface{} `json:"chat_id"`
			Text   string      `json:"text"`
		}{ChatId: chatId, Text: "BOT: " + text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}

func (client TelegramClient) SendCompanyMessage(chatId interface{}, text string) error {
	_, err := resty.R().
		SetBody(struct {
			ChatId interface{} `json:"chat_id"`
			Text   string      `json:"text"`
		}{ChatId: chatId, Text: "COMPANY: " + text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}
