package telegram

import (
	"log"

	"github.com/globalsign/mgo/bson"
)

type Bot struct {
	db       *Db
	telegram *TelegramClient
}

type DialogStatus string
type UserStatus string

const (
	DIALOG_STATUS_ACTIVE  DialogStatus = "A"
	DIALOG_STATUS_DELETED DialogStatus = "D"
	DIALOG_STATUS_REQUEST DialogStatus = "R"
)

type Dialog struct {
	ID      bson.ObjectId `json:"_id,omitempty"`
	UserA   bson.ObjectId
	AcceptA bool
	UserB   bson.ObjectId
	AcceptB bool
	Status  DialogStatus
}

type DialogRequest struct {
	ID         bson.ObjectId `json:"_id,omitempty"`
	UserId     bson.ObjectId
	Processing bool
}

type BotUser struct {
	ID           bson.ObjectId `json:"_id,omitempty"`
	TelegramID   int
	Name         string
	LanguageCode string
	Status       UserStatus
	Pause        bool
	ChatID       int64
	DialogID     bson.ObjectId
}

const (
	USER_STATUS_SEARCH        UserStatus = "S"
	USER_STATUS_COMMUNICATION UserStatus = "C"
)

func NewBot(db *Db, telegram *TelegramClient) (*Bot, error) {
	return &Bot{db: db, telegram: telegram}, nil
}

func (bot Bot) Start(user *User, chat *Chat) error {
	botUser, err := bot.db.FindUserByTelegramId(user.ID)
	if bot.db.IsNotFound(err) {
		log.Println("New User " + string(user.ID))
		botUser.TelegramID = user.ID
		botUser.Name = user.FirstName
		botUser.LanguageCode = user.LanguageCode
		botUser.ChatID = chat.ID
		botUser.Status = USER_STATUS_SEARCH
		botUser.Pause = true
	} else if err != nil {
		return err
	}

	if botUser.DialogID.Valid() {
		dialog, err := bot.db.FindDialog(botUser.DialogID)
		if bot.db.IsNotFound(err) {
			log.Println("Dialog not found: " + botUser.DialogID)
			botUser.DialogID = ""
		}
		dialogActive, err := bot.IsDialogActive(dialog)
		if err != nil {
			return err
		}
		if !dialogActive {
			log.Println("Dialog not active: " + botUser.DialogID)
			botUser.DialogID = ""
		}
	}

	err = bot.db.SaveUser(botUser)
	log.Println("User activated " + string(user.ID))
	return err
}

func (bot Bot) IsDialogActive(dialog Dialog) (bool, error) {
	if dialog.Status == DIALOG_STATUS_DELETED {
		return false, nil
	}
	userA, err := bot.db.FindUser(dialog.UserA)
	if err != nil {
		return false, err
	}
	if userA.DialogID != dialog.ID {
		return false, nil
	}
	userB, err := bot.db.FindUser(dialog.UserB)
	if err != nil {
		return false, err
	}
	if userB.DialogID != dialog.ID {
		return false, nil
	}
	return true, nil
}

func (bot Bot) Search(user *User) error {
	botUser, err := bot.db.FindUserByTelegramId(user.ID)
	if err != nil {
		return err
	}
	if botUser.DialogID != "" {
		dialog, err := bot.db.FindDialog(botUser.DialogID)
		if err != nil {
			return err
		}
		err = bot.db.DeleteDialog(dialog.ID)
		if err != nil {
			return err
		}
		var companyUserID bson.ObjectId
		if botUser.ID == dialog.UserA {
			companyUserID := dialog.UserB
		} else {
			companyUserID := dialog.UserA
		}
		bot.db.UpdateUserPause(companyUserID, true)
	}
	err = bot.db.UpdateUserStatus(botUser.ID, USER_STATUS_SEARCH)
	if err != nil {
		return err
	}
	log.Println("Start dialog request")
	err = bot.db.StartDialog(botUser.ID)
	if err != nil {
		return err
	}
	return nil
}

func (bot Bot) Pause(user *User) error {
	botUser, err := bot.db.FindUserByTelegramId(user.ID)
	if err != nil {
		return err
	}

	err = bot.db.UpdateUserPause(botUser.ID, !botUser.Pause)
	if err != nil {
		return err
	}
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
