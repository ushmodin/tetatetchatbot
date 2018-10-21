package telegram

import (
	"testing"
)

func TestStart(t *testing.T) {
	db, err := NewBoltDb()
	if err != nil {
		t.Fatal(err)
	}
	bot, err := NewBot(db, nil)
	user := User{
		ID:           42,
		FirstName:    "Ivan",
		LastName:     "Ivanov",
		UserName:     "IvanovIvan",
		LanguageCode: "RU",
		IsBot:        false,
	}
	chat := Chat{
		ID: 100500,
	}
	err = bot.Start(user, chat)
	if err != nil {
		t.Fatal(err)
	}
	err = bot.Start(user, chat)
	if err != nil {
		t.Fatal(err)
	}
}
