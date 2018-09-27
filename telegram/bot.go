package telegram

import (
	"github.com/globalsign/mgo/bson"
)

type Bot struct {
	db       *Db
	telegram *TelegramClient
}

type BotUser struct {
	ID           int
	Name         string
	LanguageCode string
	Status       string
	Pause        bool
	ChatId       int64
	DialogId     bson.ObjectId
}

const (
	USER_STATUS_SEARCH        = "S"
	USER_STATUS_COMMUNICATION = "C"
)

func NewBot(db *Db, telegram *TelegramClient) (*Bot, error) {
	return &Bot{db: db, telegram: telegram}, nil
}

func (bot Bot) Start(user *User, chat *Chat) error {
	botUser, err := bot.db.FindUser(user.ID)
	if bot.db.IsNotFound(err) {
		botUser.ID = user.ID
		botUser.Name = user.FirstName
		botUser.LanguageCode = user.LanguageCode
		botUser.ChatId = chat.ID
		botUser.Status = USER_STATUS_SEARCH
		botUser.Pause = true
		err = bot.db.SaveUser(botUser)
	}
	if err != nil {
		return err
	}

	return nil
}

func (bot Bot) Search() error {
	return nil
}

func (bot Bot) Pause() error {
	return nil
}

func (bot Bot) Status() error {
	return nil
}

func (bot Bot) Who() error {
	return nil
}

func (bot Bot) GetCurrentCompany() (interface{}, error) {
	return nil, nil
}
