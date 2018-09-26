package main

import (
	"log"
	"net/http"

	tetatet "github.com/ushmodin/tetatetchatbot/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	server, _ := tetatet.NewTetATetHttpServer()
	router.HandleFunc("/t/b/tetatet/updates", server.UpdateHandler).Methods("POST")
	router.HandleFunc("/ping", server.PingHandler).Methods("GET")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
