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
	router.HandleFunc("/t/b/tetatet/updates", client.UpdateHandler).Methods("POST")
	router.HandleFunc("/ping", client.PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
