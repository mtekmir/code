package server

import (
	"log"
	"net/http"
)

type handler func(w http.ResponseWriter, r *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// handle returned error here.
		log.Println("handle the error properly")
		w.WriteHeader(500)
		w.Write([]byte("something went wrong"))
	}
}
