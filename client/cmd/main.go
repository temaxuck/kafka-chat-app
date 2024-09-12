// TODO: Refactor whole client app
// TODO: Consider to come up with a simpler client-server conversation
//       interface

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"client/internal/utils"
	"client/pkg/app" // "github.com/temaxuck/kafka-chat-app/server/pkg/app"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "HTTP server address")
var isHost = flag.Bool("host", false, "Set host mode if you want to create a new room")

type AppState struct {
	clientId *string
	roomId   *string
}

func (s AppState) String() string {
	return fmt.Sprintf(
		"AppState{clientId: %s, roomId: %s}",
		utils.NilOrString(s.clientId),
		utils.NilOrString(s.roomId),
	)
}

func formatClientId(clientId string) string {
	return clientId[len(clientId)-6:]
}

func formatTime(timestamp time.Time) string {
	return timestamp.Format("02/01/2006 03:04 PM")
}

func Listen(c *websocket.Conn, done chan struct{}, state AppState) {
	defer close(done)

	var incoming app.Message

	for {
		err := c.ReadJSON(&incoming)

		if err != nil {
			fmt.Println("read:", err)
			return
		}

		if *incoming.SenderId != *state.clientId {
			fmt.Printf("[%6s at %s]: %s\n", formatClientId(*incoming.SenderId), formatTime(incoming.Timestamp), *incoming.Data)
		}
	}
}

func Talk(c *websocket.Conn, done chan struct{}, scanner bufio.Scanner, state AppState) {
	defer close(done)

	for {
		scanner.Scan()
		text := scanner.Text()
		if len(text) > 0 {
			err := c.WriteJSON(app.NewMessage(app.MTChatMessage, state.roomId, state.clientId, &text))
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func setupApp(
	c *websocket.Conn,
	scanner bufio.Scanner,
	state *AppState,
	done chan struct{},
) error {
	var incoming app.Message

	if *isHost {
		err := c.ReadJSON(&incoming)
		if err != nil {
			return err
		}
		if incoming.Type == app.MTInit {
			state.clientId = incoming.Data
			state.roomId = incoming.RoomId
		}
		fmt.Printf("Room has been created! Room id: %s\n", *state.roomId)
		go Listen(c, done, *state)
		go Talk(c, done, scanner, *state)
		return nil
	}

	for {
		fmt.Print("Join by room id: ")
		scanner.Scan()
		text := scanner.Text()
		if len(text) > 0 {
			err := c.WriteJSON(app.NewMessage(app.MTInit, &text, nil, nil))
			if err != nil {
				return err
			}
		}
		err := c.ReadJSON(&incoming)
		if err != nil {
			return err
		}

		if incoming.Type == app.MTInit {
			fmt.Printf("Successfully connected to room#%s\n", text)
			break
		}

		fmt.Println(string(incoming.Type), *incoming.Data)

		fmt.Printf("Room#%s doesn't exist. Please try again.\n", text)
	}

	state.clientId = incoming.Data
	state.roomId = incoming.RoomId

	go Listen(c, done, *state)
	go Talk(c, done, scanner, *state)

	return nil
}

func main() {
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/join"}
	if *isHost {
		u.Path = "/host"
	}

	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	scanner := bufio.NewScanner(os.Stdin)

	var state AppState

	go setupApp(c, *scanner, &state, done)

	// Handle signals
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)

			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
