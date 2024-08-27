package middleware

import (
	"log"
	"net/http"
)

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recieved request from \"%s\": %s %s\n", r.RemoteAddr, r.Method, r.URL)

		h.ServeHTTP(w, r)
	})
}
