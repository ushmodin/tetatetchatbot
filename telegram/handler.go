package telegram

import (
	"encoding/json"
	"log"
	"net/http"
)

type HTTPHandler struct {
	bot      *Bot
	telegram *TelegramClient
}

type UserError struct {
	message string
}

func (err UserError) Error() string {
	return err.message
}

func NewUserError(message string) error {
	return &UserError{message: message}
}

func NewHTTPHandler(bot *Bot, telegram *TelegramClient) (*HTTPHandler, error) {
	return &HTTPHandler{bot, telegram}, nil
}

func (handler HTTPHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var update Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Println(err)
	}

	if update.Message.Text == "" {
		return
	}
	log.Println("Command " + update.Message.Text)
	if update.Message.Text[:1] == "/" {
		cmd := update.Message.Text[1:]
		if cmd == "start" {
			handler.bot.Start(*update.Message.From, *update.Message.Chat)
			return
		} else if cmd == "search" {
			handler.bot.Search(update.Message.From)
			return
		} else if cmd == "pause" {
			handler.bot.Pause(update.Message.From)
			return
		} else if cmd == "status" {
			handler.bot.Status()
			return
		} else if cmd == "who" {
			handler.bot.Who()
			return
		} else {
			err = handler.telegram.SendMessage(update.Message.Chat.ID, "BOT: Unknow command "+cmd)
			if err != nil {
				log.Println(err)
			}
		}
	}

	chatID, err := handler.bot.GetCurrentCompany(update.Message.From)
	message := update.Message.Text
	if _, ok := err.(*UserError); ok {
		chatID = update.Message.Chat.ID
		message = err.Error()
	}
	err = handler.telegram.SendMessage(chatID, message)
	if err != nil {
		log.Println(err)
	}
}

func (handler HTTPHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	w.Header()["Content-type"] = []string{"text/plain"}
}
