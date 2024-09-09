package server

import (
	"log"
	"net/http"
	"server/internal/middleware"
	"server/internal/routes"
)

func handle(pattern string, h func(http.ResponseWriter, *http.Request)) {
	http.Handle(
		pattern,
		middleware.WithLogging(http.HandlerFunc(h)),
	)
}

func Run(addr *string) {
	log.Printf("Starting http server: %s\n", *addr)

	handle("/", routes.Home)
	handle("/echo", routes.Echo)
	handle("/host", routes.Host)
	handle("/join", routes.Join)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
