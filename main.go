package main

import (
	"log"
	"net/http"

	"github.com/ushmodin/tetatetchatbot/telegram"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	client, _ := telegram.NewTelegramClient("")
	router.HandleFunc("/t/b/tetatet/updates", client.UpdateHandler).Methods("POST")
	router.HandleFunc("/ping", client.PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
