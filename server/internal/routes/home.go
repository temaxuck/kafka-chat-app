package routes

import (
	"net/http"
	"server/internal/constants"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to chat app!\n" + constants.USAGE_MESSAGE))
}
