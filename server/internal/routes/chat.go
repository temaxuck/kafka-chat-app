package routes

import (
	"fmt"
	"log"
	"net/http"
	"server/pkg/app"
	"sync"
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

	initMsg := app.NewMessage(app.MTInit, &room.RoomId, nil, &client.ClientId)
	err = c.WriteJSON(initMsg)
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

	// Prompt for room id
	for {
		var connectMsg app.Message
		err = c.ReadJSON(&connectMsg)
		if err != nil {
			log.Println("deserialize:", err)
			sErr := fmt.Sprintf("Couldn't read json: %v", err)
			errorMsg := app.NewMessage(app.MTError, nil, nil, &sErr)
			err = c.WriteJSON(errorMsg)
			continue
		}

		if connectMsg.RoomId == nil {
			roomId = ""
		} else {
			roomId = *connectMsg.RoomId
		}

		if _, ok := rooms[roomId]; ok {
			break
		}

		sErr := "Wrong room id, try again please"
		errorMsg := app.NewMessage(app.MTError, nil, nil, &sErr)
		err = c.WriteJSON(errorMsg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}

	log.Printf("Client %s has been connected to room %s\n", c.RemoteAddr(), roomId)
	room := rooms[roomId]
	client := app.NewClient(&room, c, kafkaServers)
	room.AddClient(client)

	initMsg := app.NewMessage(app.MTInit, &room.RoomId, nil, &client.ClientId)
	err = c.WriteJSON(initMsg)
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
