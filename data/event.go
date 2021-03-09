package data

import (
	proto "github.com/golang/protobuf/proto"
)

// FromBinary loads an event from a byte array
func (event *Event) FromBinary(msg []byte) error {
	return proto.Unmarshal(msg, event)
}

// ToBinary converts an event to a byte string
func (event *Event) ToBinary() ([]byte, error) {
	return proto.Marshal(event)
}
