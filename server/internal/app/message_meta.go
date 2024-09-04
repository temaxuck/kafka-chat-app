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
