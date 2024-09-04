package server

import (
	"fmt"
	"log"
	"net/http"
	"server/internal/app"
	"server/internal/constants"
	"server/internal/middleware"
	"sync"

	"github.com/gorilla/websocket"
)

const kafkaServers = "localhost:29092" // TODO: Unhardcode

var upgrader = websocket.Upgrader{}

// TODO: Track rooms and dispose unused
var rooms = make(map[string]app.Room)

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

// Host - Create room with id and subscribe listener to kafka topic with key id
func host(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return

	}
	defer c.Close()

	// Create room
	room := app.NewRoom()
	rooms[room.RoomId] = room

	// Create client
	client := app.NewClient(&room, c, kafkaServers)
	room.AddClient(client)

	log.Printf("Room#%s has been created and is hosted by Client#%s\n", room.RoomId, client.ClientId)

	err = c.WriteMessage(
		websocket.TextMessage,
		[]byte(
			fmt.Sprintf("You are hosting room with id: %s\nShare the id with someone to have a chat!", room.RoomId),
		),
	)
	if err != nil {
		log.Println("write:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go client.Listen()
	go client.Talk(&wg)

	wg.Wait()

	room.RemoveClient(client)
}

// Join - Join room by id and subscribe listener to kafka topic with key id
func join(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return

	}
	defer c.Close()

	var roomId string

	for {
		err = c.WriteMessage(
			websocket.TextMessage,
			[]byte(
				"Join chat using room id: ",
			),
		)
		if err != nil {
			log.Println("write:", err)
			return
		}

		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		roomId = string(msg)

		if _, ok := rooms[roomId]; ok {
			break
		}

		err = c.WriteMessage(
			websocket.TextMessage,
			[]byte(
				"Wrong room id, try again please: ",
			),
		)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}

	log.Printf("Client %s has been connected to room %s\n", c.RemoteAddr(), roomId)
	room := rooms[roomId]
	client := app.NewClient(&room, c, kafkaServers)
	room.AddClient(client)

	err = c.WriteMessage(
		websocket.TextMessage,
		[]byte(
			fmt.Sprintf("You successfully connected to room with id: %s", room.RoomId),
		),
	)
	if err != nil {
		log.Println("write:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go client.Listen()
	go client.Talk(&wg)

	wg.Wait()

	room.RemoveClient(client)
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
	Handle("/host", host)
	Handle("/join", join)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
