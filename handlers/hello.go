package handlers

import (
	"net/http"
)

func HandlerHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}
