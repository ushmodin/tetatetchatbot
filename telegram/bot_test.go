package telegram

import (
	"fmt"
	"log"
	"os"
	"testing"
)

var db *BoltDb
var ms *BoltMessageService
var bot *Bot

func setUp() error {
	var err error
	db, err = NewBoltDb()
	if err != nil {
		return err
	}
	ms, err = NewBoltMessageService()
	if err != nil {
		return err
	}
	bot, err = NewBot(db, ms)
	return err
}

func tearDown() error {
	defer db.Close()
	defer ms.Close()
	os.Remove("bms.db")
	os.Remove("tetatetchatbot.db")
	return nil
}

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		log.Panic(err)
	}
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

func TestStart(t *testing.T) {
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
	err := bot.Start(user, chat)
	if err != nil {
		t.Fatal(err)
	}
	err = bot.Start(user, chat)
	if err != nil {
		t.Fatal(err)
	}

	if ms == nil {
		log.Panic("Service is nil2")
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
