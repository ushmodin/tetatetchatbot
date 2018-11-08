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
	BackwardRequestDialog(dlgReq DialogRequest) error
}

type MessageService interface {
	SendServiceMessage(chatId int64, text string) error
	SendCompanyMessage(chatId int64, text string) error
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
	Created    int64         `bson:"Created"`
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
	if err = bot.messageService.SendServiceMessage(chat.ID, "Hello, "+user.UserName); err != nil {
		return err
	}
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

func (bot Bot) Search(user User) error {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if err != nil {
		return err
	}
	if botUser.Status == USER_STATUS_SEARCH && !botUser.Pause {
		return bot.messageService.SendServiceMessage(botUser.ChatID, "I'm still search")
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
	err = bot.db.UpdateUserPause(botUser.ID, false)
	if err != nil {
		return err
	}
	return nil
}

func (bot Bot) createDialog(reqA, reqB DialogRequest) error {
	dialogID, err := bot.db.CreateDialog(reqA, reqB)
	if err != nil {
		bot.db.BackwardRequestDialog(reqA)
		bot.db.BackwardRequestDialog(reqB)
		return err
	}
	err = bot.db.UpdateUserDialog(reqA.UserID, &dialogID)
	if err != nil {
		bot.db.BackwardRequestDialog(reqA)
		bot.db.BackwardRequestDialog(reqB)
		return err
	}
	err = bot.db.UpdateUserDialog(reqB.UserID, &dialogID)
	if err != nil {
		bot.db.BackwardRequestDialog(reqA)
		bot.db.BackwardRequestDialog(reqB)
		return err
	}
	log.Printf("Dialog for %s and %s created", reqA.UserID, reqB.UserID)
	userA, err := bot.db.FindUser(reqA.UserID)
	if err != nil {
		return err
	}
	userB, err := bot.db.FindUser(reqB.UserID)
	if err != nil {
		return err
	}
	if err := bot.messageService.SendServiceMessage(userA.ChatID, "I found company for you. Dialog started"); err != nil {
		return err
	}
	return bot.messageService.SendServiceMessage(userB.ChatID, "I found company for you. Dialog started")
}

func (bot Bot) findNextDialogRequest() (DialogRequest, error) {
	for {
		req, err := bot.db.FindNextDialogRequest()
		if err != nil {
			return DialogRequest{}, err
		}
		user, err := bot.db.FindUser(req.UserID)
		if err != nil {
			return DialogRequest{}, err
		}
		if user.Status != USER_STATUS_SEARCH {
			continue
		}
		return req, nil
	}

}

func (bot Bot) JoinRequests() (bool, error) {
	reqA, err := bot.findNextDialogRequest()
	if bot.db.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	log.Println("Request A found " + reqA.ID)
	reqB, err := bot.findNextDialogRequest()
	if bot.db.IsNotFound(err) {
		bot.db.BackwardRequestDialog(reqA)
		return false, nil
	} else if err != nil {
		bot.db.BackwardRequestDialog(reqA)
		return false, err
	}
	log.Println("Request B found " + reqB.ID)
	err = bot.createDialog(reqA, reqB)
	return true, err
}

func (bot Bot) JoinRequestsLoop() {
	for {
		ok, err := bot.JoinRequests()
		if err != nil {
			log.Print(err)
		}
		if !ok {
			time.Sleep(1 * time.Second)
		}
	}
}
func (bot Bot) Pause(user User) error {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if err != nil {
		return err
	}

	value := !botUser.Pause
	err = bot.db.UpdateUserPause(botUser.ID, value)
	if err != nil {
		return err
	}

	if value {
		companyChatID, err := bot.GetCurrentCompany(user)
		if err != nil {
			return err
		}
		if companyChatID != 0 {
			if err := bot.messageService.SendServiceMessage(companyChatID, "You company go to pause"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bot Bot) Status() error {
	return nil
}

func (bot Bot) Who() error {
	return nil
}

func (bot Bot) GetCurrentCompany(user User) (int64, error) {
	botUser, err := bot.db.FindUserByTelegramID(user.ID)
	if err != nil {
		return 0, err
	}
	if botUser.DialogID == nil {
		return 0, nil
	}
	dialog, err := bot.db.FindDialog(*botUser.DialogID)
	if err != nil {
		return 0, err
	}
	if dialog.Status != DIALOG_STATUS_ACTIVE {
		bot.db.UpdateUserDialog(botUser.ID, nil)
		bot.db.UpdateUserPause(botUser.ID, true)
		return 0, nil
	}
	var companyUserID bson.ObjectId
	if dialog.UserA == botUser.ID {
		companyUserID = dialog.UserB
	} else {
		companyUserID = dialog.UserA
	}
	companyUser, err := bot.db.FindUser(companyUserID)
	if err != nil {
		return 0, err
	}
	return companyUser.ChatID, nil
}
