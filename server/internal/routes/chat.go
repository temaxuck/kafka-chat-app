package routes

import (
	"fmt"
	"log"
	"net/http"
	"server/internal/app"
	"sync"

	"github.com/gorilla/websocket"
)

const kafkaServers = "localhost:29092" // TODO: Unhardcode

var rooms = make(map[string]app.Room)

// Creates Room and subscribes Client to it
func Host(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)
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

	if room.RemoveClient(client) {
		delete(rooms, room.RoomId)
		log.Printf("Room#%s has been freed. Active rooms: %d\n", room.RoomId, len(rooms))
	}
}

// Subscribes Client to Room with RoomID
func Join(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)
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

	if room.RemoveClient(client) {
		delete(rooms, room.RoomId)
		log.Printf("Room#%s has been freed. Active rooms: %d\n", room.RoomId, len(rooms))
	}
}
