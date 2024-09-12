package app

import (
	"encoding/json"
	"errors"
	"time"
)

type MessageType string

const (
	MTInfo        MessageType = "info"
	MTInit        MessageType = "init"
	MTError       MessageType = "error"
	MTChatMessage MessageType = "chat_message"
	MTNoMatch     MessageType = ""
)

type Message struct {
	// If SenderId is nil, message is sent by server
	Type      MessageType `json:"type"`
	RoomId    *string     `json:"room_id"`
	SenderId  *string     `json:"sender_id"`
	Data      *string     `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewMessage(msgType MessageType, roomId, senderId, data *string) Message {
	return Message{
		Type:      msgType,
		RoomId:    roomId,
		SenderId:  senderId,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func (m *Message) UnmarshalJSON(d []byte) error {
	type Aux Message
	var a *Aux = (*Aux)(m)
	err := json.Unmarshal(d, &a)
	if err != nil {
		return err
	}

	switch m.Type {
	case MTInfo, MTInit, MTError, MTChatMessage:
		return nil
	default:
		m.Type = MTNoMatch
		return errors.New("Couldn't deserialize message type")
	}
}
