package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"server/internal/constants"
	"server/internal/middleware"
)

var upgrader = websocket.Upgrader{}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to chat app!\n" + constants.USAGE_MESSAGE))
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return

	}
	defer c.Close()

	for {
		t, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("Recieved message from %s: %s\n", r.RemoteAddr, msg)

		err = c.WriteMessage(t, msg)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}

}

func Handle(pattern string, h func(http.ResponseWriter, *http.Request)) {
	http.Handle(
		pattern,
		middleware.WithLogging(http.HandlerFunc(h)),
	)
}

func Run(addr *string) {
	log.Printf("Starting http server: %s\n", *addr)

	Handle("/", home)
	Handle("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
