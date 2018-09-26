package main

import (
	"net/http"

	tetatet "github.com/ushmodin/tetatetchatbot/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/t/b/tetatet/updates", tetatet.TetATetBotHttpUpdate).Methods("POST")
	router.HandleFunc("/ping", tetatet.Ping).Methods("GET")
	http.ListenAndServe("localhost:8080", router)
}
