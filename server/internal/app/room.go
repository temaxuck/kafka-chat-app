package app

import (
	"server/internal/constants"
	"server/internal/utils"
)

type Room struct {
	RoomId  string
	clients map[string]Client
}

func NewRoom() Room {
	roomId, err := utils.GenerateRoomId(constants.ROOM_ID_LENGTH)
	if err != nil {
		panic(err)
	}
	return Room{
		RoomId:  roomId,
		clients: make(map[string]Client),
	}
}

func (r *Room) AddClient(c Client) {
	r.clients[c.ClientId] = c
}

func (r *Room) RemoveClient(c Client) {
	delete(r.clients, c.ClientId)
}
