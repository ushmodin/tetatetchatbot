package tetatet

import (
	"io"
	"net/http"
)

type TetATetHttpServer struct {
}

func NewTetATetHttpServer() (*TetATetHttpServer, error) {
	return &TetATetHttpServer{}, nil
}

func (server TetATetHttpServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
}

func (server TetATetHttpServer) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	w.Header()["Content-type"] = []string{"text/plain"}
}
