package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ushmodin/tetatetchatbot/telegram"
)

func main() {
	router := mux.NewRouter()
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Telegram token not specified")
	}
	client, _ := telegram.NewTelegramClient(token)
	db, err := telegram.NewDb("mongodb:27017", "tetatetchatbot")
	if err != nil {
		log.Fatal(err)
	}
	bot, _ := telegram.NewBot(db, client)
	handler, _ := telegram.NewHTTPHandler(bot, client)
	router.HandleFunc("/t/b/tetatet/updates", handler.UpdateHandler).Methods("POST")
	router.HandleFunc("/ping", handler.PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
