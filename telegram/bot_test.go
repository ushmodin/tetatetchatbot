package telegram

import (
	"fmt"
	"testing"
)

func TestStart(t *testing.T) {
	db, err := NewBoltDb()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ms, err := NewBoltMessageService()
	if err != nil {
		t.Fatal(err)
	}
	defer ms.Close()
	bot, err := NewBot(db, ms)
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
	messages, err := ms.Next10()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(messages))
	if len(messages) != 2 {
		t.Fatal("Incorrect message count")
	}
	if messages[0].ChatID != 100500 {
		t.Fatal("Incorrect response message")
	}
}
