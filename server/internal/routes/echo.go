package routes

import (
	"log"
	"net/http"
)

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)
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
