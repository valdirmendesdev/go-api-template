package rest

import (
	"net/http"
)

func (s Server) WelcomeMessage(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Welcome to the REST API"))
}
