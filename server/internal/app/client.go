package app

import (
	"fmt"
	"log"
	"server/internal/constants"
	"server/internal/utils"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"
)

type Client struct {
	alive    bool
	room     *Room
	ClientId string
	wsConn   *websocket.Conn
	consumer *kafka.Consumer
	producer *kafka.Producer
}

func NewClient(room *Room, wsConn *websocket.Conn, kafkaServers string) Client {
	c := Client{
		alive:    true,
		room:     room,
		ClientId: utils.GenerateUUID(),
		wsConn:   wsConn,
	}
	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServers,
		"group.id":          c.ClientId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}
	log.Printf("Registered a new consumer with group id: %s\n", c.ClientId)

	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServers,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
	c.consumer = cons
	c.producer = prod

	return c
}

// Subscribe to chat room (kafka) and forward recieved messages to ws connection
func (c *Client) Listen() {
	topic := constants.CHAT_TOPIC

	defer c.consumer.Close()

	err := c.consumer.Assign([]kafka.TopicPartition{
		{Topic: &topic, Partition: 0},
	})

	if err != nil {
		panic(err)
	}

	for c.alive {
		incoming, err := c.consumer.ReadMessage(10 * time.Millisecond)
		if err == nil {
			var metadata MessageMeta
			err := metadata.deserialize(incoming.Key)
			if err != nil {
				log.Println("Couldn't deserialize message metadata:", err)
				continue
			}
			if metadata.RoomId == c.room.RoomId && metadata.SenderId != c.ClientId {
				err = c.wsConn.WriteMessage(websocket.TextMessage, incoming.Value)
				if err != nil {
					log.Println("write:", err)
					break
				}
			}
		} else if !err.(kafka.Error).IsTimeout() {
			log.Printf("Consumer error: %v (%v)\n", err, incoming)
		}
	}
}

// Forward messages sent from ws connection to chat room (kafka)
func (c *Client) Talk(wg *sync.WaitGroup) {
	topic := constants.CHAT_TOPIC

	defer c.producer.Close()

	for {
		_, outcoming, err := c.wsConn.ReadMessage()
		if err != nil {
			// Client disconnected
			wg.Done() // is needed to notify server that ws connection has been closed
			c.alive = false
			log.Println("read:", err)
			break
		}

		log.Printf("Recieved message from %s: %s\n", c.wsConn.RemoteAddr(), outcoming)
		metadata, err := (MessageMeta{RoomId: c.room.RoomId, SenderId: c.ClientId}).serialize()
		if err != nil {
			log.Println("Message hasn't been sent. Couldn't serialize message metadata:", err)
			continue
		}
		c.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
			},

			Key:   metadata,
			Value: []byte(outcoming),
		}, nil)
	}
}
