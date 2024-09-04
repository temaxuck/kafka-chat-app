package app

import (
	"encoding/json"
)

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

// func listenAndWriteRoom(c *websocket.Conn, roomId string) {
// 	topic := "chat"

// 	prod, err := kafka.NewProducer(&kafka.ConfigMap{
// 		"bootstrap.servers": "localhost:29092",
// 	})
// 	if err != nil {
// 		panic(fmt.Sprintf("Failed to create producer: %v", err))
// 	}
// 	defer prod.Close()
// 	uuid := utils.GenerateUUID()

// 	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
// 		"bootstrap.servers": "localhost:29092",
// 		"group.id":          uuid,
// 		"auto.offset.reset": "earliest",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer cons.Close()

// 	log.Printf("Registered a new consumer with group id: %s\n", uuid)

// 	err = cons.Assign([]kafka.TopicPartition{
// 		{Topic: &topic, Partition: 0},
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	go func() {
// 		for e := range prod.Events() {
// 			switch ev := e.(type) {
// 			case *kafka.Message:
// 				if ev.TopicPartition.Error != nil {
// 					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
// 				} else {
// 					log.Printf("Delivered message to %v\n", ev.TopicPartition)
// 				}
// 			}
// 		}
// 	}()

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	run := true

// 	go func() {
// 		for {
// 			_, outcoming, err := c.ReadMessage()
// 			if err != nil {
// 				// Client disconnected
// 				wg.Done()
// 				run = false
// 				log.Println("read:", err)
// 				break
// 			}

// 			log.Printf("Recieved message from %s: %s\n", c.RemoteAddr(), outcoming)
// 			metadata, err := (MessageMeta{RoomId: roomId, SenderId: uuid}).serialize()
// 			if err != nil {
// 				log.Println("Message hasn't been sent. Couldn't serialize message metadata:", err)
// 				continue
// 			}
// 			prod.Produce(&kafka.Message{
// 				TopicPartition: kafka.TopicPartition{
// 					Topic:     &topic,
// 					Partition: 0,
// 				},

// 				Key:   metadata,
// 				Value: []byte(outcoming),
// 			}, nil)
// 		}
// 	}()

// 	go func() {
// 		for run {
// 			incoming, err := cons.ReadMessage(10 * time.Millisecond)
// 			if err == nil {
// 				var metadata MessageMeta
// 				err := metadata.deserialize(incoming.Key)
// 				if err != nil {
// 					log.Println("Couldn't deserialize message metadata:", err)
// 					continue
// 				}
// 				if metadata.RoomId == roomId && metadata.SenderId != uuid {
// 					err = c.WriteMessage(websocket.TextMessage, incoming.Value)
// 					if err != nil {
// 						log.Println("write:", err)
// 						break
// 					}
// 				}
// 			} else if !err.(kafka.Error).IsTimeout() {
// 				log.Printf("Consumer error: %v (%v)\n", err, incoming)
// 			}
// 		}

// 	}()

// 	wg.Wait()
// }
