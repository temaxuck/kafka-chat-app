package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"

	"server/internal/constants"
	"server/internal/middleware"
	"server/internal/utils"
)

const ID_LENGTH = 8

var upgrader = websocket.Upgrader{}
var roomIds = []string{}

type MessageMeta struct {
	RoomId   string
	SenderId string
}

func (m MessageMeta) serialize() ([]byte, error) {
	d, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	return d, nil
}

func (m *MessageMeta) deserialize(d []byte) error {
	return json.Unmarshal(d, m)
}

func listenAndWriteRoom(c *websocket.Conn, roomId string) {
	topic := "chat"

	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:29092",
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
	defer prod.Close()
	uuid := utils.GenerateUUID()

	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:29092",
		"group.id":          uuid,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}
	defer cons.Close()

	log.Printf("Registered a new consumer with group id: %s\n", uuid)

	err = cons.Assign([]kafka.TopicPartition{
		{Topic: &topic, Partition: 0},
	})

	if err != nil {
		panic(err)
	}

	go func() {
		for e := range prod.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	run := true

	go func() {
		for {
			_, outcoming, err := c.ReadMessage()
			if err != nil {
				// Client disconnected
				wg.Done()
				run = false
				log.Println("read:", err)
				break
			}

			log.Printf("Recieved message from %s: %s\n", c.RemoteAddr(), outcoming)
			metadata, err := (MessageMeta{RoomId: roomId, SenderId: uuid}).serialize()
			if err != nil {
				log.Println("Message hasn't been sent. Couldn't serialize message metadata:", err)
				continue
			}
			prod.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0,
				},

				Key:   metadata,
				Value: []byte(outcoming),
			}, nil)
		}
	}()

	go func() {
		for run {
			incoming, err := cons.ReadMessage(10 * time.Millisecond)
			if err == nil {
				var metadata MessageMeta
				err := metadata.deserialize(incoming.Key)
				if err != nil {
					log.Println("Couldn't deserialize message metadata:", err)
					continue
				}
				if metadata.RoomId == roomId && metadata.SenderId != uuid {
					err = c.WriteMessage(websocket.TextMessage, incoming.Value)
					if err != nil {
						log.Println("write:", err)
						break
					}
				}
			} else if !err.(kafka.Error).IsTimeout() {
				log.Printf("Consumer error: %v (%v)\n", err, incoming)
			}
		}

	}()

	wg.Wait()
}

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

	roomId, err := utils.GenerateRoomId(ID_LENGTH)
	if err != nil {
		log.Println("Couldn't generate id:", err)
		return
	}

	roomIds = append(roomIds, roomId)

	log.Printf("Room#%s has been created and is hosted by %s\n", roomId, c.RemoteAddr())

	err = c.WriteMessage(
		websocket.TextMessage,
		[]byte(
			fmt.Sprintf("You are hosting room with id: %s\nShare the id with someone to have a chat!", roomId),
		),
	)
	if err != nil {
		log.Println("write:", err)
		return
	}

	listenAndWriteRoom(c, roomId)
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

		if slices.Contains(roomIds, roomId) {
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
	listenAndWriteRoom(c, roomId)
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
