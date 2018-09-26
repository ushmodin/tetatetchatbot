package tetatet

import (
	"io"
	"net/http"

	"github.com/ushmodin/tetatetchatbot/telegram"
)

type TetATetHttpServer struct {
	telegram *telegram.TelegramClient
}

func NewTetATetHttpServer(telegram *telegram.TelegramClient) (*TetATetHttpServer, error) {
	return &TetATetHttpServer{telegram: telegram}, nil
}

func (server TetATetHttpServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
	server.telegram.SendMessage()
}

func (server TetATetHttpServer) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	w.Header()["Content-type"] = []string{"text/plain"}
}
