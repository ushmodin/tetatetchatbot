package tetatet

import (
	"io"
	"net/http"
)

func TetATetBotHttpUpdate(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	w.Header()["Content-type"] = []string{"text/plain"}
}
