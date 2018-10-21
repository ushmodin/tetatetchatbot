package telegram

import (
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
)

type Db interface {
	FindUserByTelegramID(id int) (BotUser, error)
	IsNotFound(err error) bool
	FindUser(id bson.ObjectId) (BotUser, error)
	SaveUser(user BotUser) error
	FindDialog(id bson.ObjectId) (Dialog, error)
	DeleteDialog(id bson.ObjectId) error
	UpdateUserStatus(userID bson.ObjectId, status UserStatus) error
	UpdateUserPause(userID bson.ObjectId, flag bool) error
	UpdateUserDialog(userID bson.ObjectId, dialogID *bson.ObjectId) error
	StartDialog(userID bson.ObjectId) error
	FindNextDialogRequest() (DialogRequest, error)
	CreateDialog(reqA DialogRequest, reqB DialogRequest) (bson.ObjectId, error)
	UpdateDialogRequestProcessing(id bson.ObjectId, processing bool) error
}

type MessageService interface {
	SendServiceMessage(chatId interface{}, text string) error
	SendCompanyMessage(chatId interface{}, text string) error
}

type Bot struct {
	db             Db
	messageService MessageService
}

type DialogStatus string
type UserStatus string

const (
	DIALOG_STATUS_ACTIVE  DialogStatus = "A"
	DIALOG_STATUS_DELETED DialogStatus = "D"
	DIALOG_STATUS_REQUEST DialogStatus = "R"
)

type Dialog struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	UserA   bson.ObjectId `bson:"UserA"`
	AcceptA bool          `bson:"AcceptA"`
	UserB   bson.ObjectId `bson:"UserB"`
	AcceptB bool          `bson:"AcceptB"`
	Status  DialogStatus  `bson:"Status"`
}

type DialogRequest struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	UserID     bson.ObjectId `bson:"UserId"`
	Processing bool          `bson:"Processing"`
}

type BotUser struct {
	ID           bson.ObjectId  `bson:"_id,omitempty"`
	TelegramID   int            `bson:"TelegramID"`
	Name         string         `bson:"Name"`
	LanguageCode string         `bson:"LanguageCode"`
	Status       UserStatus     `bson:"Status"`
	Pause        bool           `bson:"Pause"`
	ChatID       int64          `bson:"ChatID"`
	DialogID     *bson.ObjectId `bson:"DialogID"`
}

const (
	USER_STATUS_SEARCH        UserStatus = "S"
	USER_STATUS_COMMUNICATION UserStatus = "C"
)

func NewBot(db Db, messageService MessageService) (*Bot, error) {
	return &Bot{db: db, messageService: messageService}, nil
}

func (bot Bot) Start(user User, chat Chat) error {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if bot.db.IsNotFound(err) {
		log.Printf("New User %d", user.ID)
		botUser.ID = bson.NewObjectId()
		botUser.TelegramID = user.ID
		botUser.Name = user.FirstName
		botUser.LanguageCode = user.LanguageCode
		botUser.ChatID = chat.ID
		botUser.Status = USER_STATUS_SEARCH
		botUser.Pause = true
		err = bot.db.SaveUser(botUser)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if botUser.DialogID != nil {
		dialogID := *botUser.DialogID
		dialog, err := bot.db.FindDialog(dialogID)
		if bot.db.IsNotFound(err) {
			bot.db.UpdateUserDialog(botUser.ID, nil)
			bot.db.UpdateUserPause(botUser.ID, true)
		}
		dialogActive, err := bot.IsDialogActive(dialog)
		if err != nil {
			return err
		}
		if !dialogActive {
			bot.db.UpdateUserDialog(botUser.ID, nil)
			bot.db.UpdateUserPause(botUser.ID, true)
		}
	}

	log.Printf("User activated %d", user.ID)
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
	if *userA.DialogID != dialog.ID {
		return false, nil
	}
	userB, err := bot.db.FindUser(dialog.UserB)
	if err != nil {
		return false, err
	}
	if *userB.DialogID != dialog.ID {
		return false, nil
	}
	return true, nil
}

func (bot Bot) Search(user *User) error {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if err != nil {
		return err
	}
	log.Printf("User go to search mode %d", user.ID)
	if botUser.DialogID != nil {
		dialog, err := bot.db.FindDialog(*botUser.DialogID)
		if err != nil {
			return err
		}
		err = bot.db.DeleteDialog(dialog.ID)
		if err != nil {
			return err
		}
		log.Printf("Dialog %s marked as deleted", dialog.ID)
		var companyUserID bson.ObjectId
		if botUser.ID == dialog.UserA {
			companyUserID = dialog.UserB
		} else {
			companyUserID = dialog.UserA
		}
		err = bot.db.UpdateUserPause(companyUserID, true)
		if err != nil {
			return err
		}
		log.Printf("User %s go to pause mode", companyUserID)
	}
	err = bot.db.UpdateUserStatus(botUser.ID, USER_STATUS_SEARCH)
	if err != nil {
		return err
	}
	log.Printf("Start dialog request for user %s", botUser.ID)
	err = bot.db.StartDialog(botUser.ID)
	if err != nil {
		return err
	}
	return nil
}

func (bot Bot) JoinRequests() error {
	for {
		reqA, err := bot.db.FindNextDialogRequest()
		if bot.db.IsNotFound(err) {
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return nil
		}
		log.Println("Request A found " + reqA.ID)

		go func() {
			var reqB DialogRequest
			for {
				reqB, err = bot.db.FindNextDialogRequest()
				if bot.db.IsNotFound(err) {
					time.Sleep(1 * time.Second)
					continue
				} else if err != nil {
					bot.db.UpdateDialogRequestProcessing(reqA.ID, false)
					return
				}
				log.Println("Request B found " + reqA.ID)
				break
			}
			dialogId, err := bot.db.CreateDialog(reqA, reqB)
			if err != nil {
				return
			}
			err = bot.db.UpdateUserDialog(reqA.UserID, &dialogId)
			if err != nil {
				return
			}
			err = bot.db.UpdateUserDialog(reqB.UserID, &dialogId)
			if err != nil {
				return
			}
			log.Println("Dialog created")
		}()
		break
	}
	return nil
}

func (bot Bot) Pause(user *User) error {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
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

func (bot Bot) GetCurrentCompany(user *User) (interface{}, error) {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if err != nil {
		return nil, err
	}
	dialog, err := bot.db.FindDialog(*botUser.DialogID)
	if err != nil {
		return nil, err
	}
	if dialog.Status != DIALOG_STATUS_ACTIVE {
		bot.db.UpdateUserDialog(botUser.ID, nil)
		bot.db.UpdateUserPause(botUser.ID, true)
		return nil, NewUserError("Dialog was interupted")
	}
	var companyUserID bson.ObjectId
	if dialog.UserA == botUser.ID {
		companyUserID = dialog.UserB
	} else {
		companyUserID = dialog.UserA
	}
	companyUser, err := bot.db.FindUser(companyUserID)
	if err != nil {
		return nil, err
	}
	return companyUser.ChatID, nil
}
