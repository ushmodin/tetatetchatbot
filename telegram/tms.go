package telegram

import (
	"gopkg.in/resty.v1"
)

type telegramClient struct {
	appKey string
}

func NewTelegramClient(appKey string) (*telegramClient, error) {
	return &telegramClient{appKey: appKey}, nil
}

func (client telegramClient) SendServiceMessage(chatId int64, text string) error {
	_, err := resty.R().
		SetBody(struct {
			ChatId interface{} `json:"chat_id"`
			Text   string      `json:"text"`
		}{ChatId: chatId, Text: "BOT: " + text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}

func (client telegramClient) SendCompanyMessage(chatId int64, text string) error {
	_, err := resty.R().
		SetBody(struct {
			ChatId interface{} `json:"chat_id"`
			Text   string      `json:"text"`
		}{ChatId: chatId, Text: "COMPANY: " + text}).
		Post("https://api.telegram.org/bot" + client.appKey + "/sendMessage")
	return err
}
