package emitter

import (
	"github.com/zishang520/socket.io/socket"
)

const UID = "emitter"

type RequestType byte

const (
	SOCKETS           RequestType = '0'
	ALL_ROOMS         RequestType = '1'
	REMOTE_JOIN       RequestType = '2'
	REMOTE_LEAVE      RequestType = '3'
	REMOTE_DISCONNECT RequestType = '4'
	REMOTE_FETCH      RequestType = '5'
	SERVER_SIDE_EMIT  RequestType = '6'
)

type Parser interface {
	Encoder(any) ([]byte, error)
}

type BroadcastOptions struct {
	Nsp              string
	BroadcastChannel string
	RequestChannel   string
	Parser           Parser
}

type BroadcastFlags struct {
	Volatile bool `json:"volatile,omitempty" msgpack:"volatile,omitempty"`
	Compress bool `json:"compress,omitempty" msgpack:"compress,omitempty"`
}

type PacketOptions struct {
	Rooms  []socket.Room   `json:"rooms,omitempty" msgpack:"rooms,omitempty"`
	Except []socket.Room   `json:"except,omitempty" msgpack:"except,omitempty"`
	Flags  *BroadcastFlags `json:"flags,omitempty" msgpack:"flags,omitempty"`
}

type Packet struct {
	Uid   string         `json:"uid,omitempty" msgpack:"uid,omitempty"`
	Type  RequestType    `json:"type,omitempty" msgpack:"type,omitempty"`
	Data  any            `json:"data,omitempty" msgpack:"data,omitempty"`
	Opts  *PacketOptions `json:"opts,omitempty" msgpack:"opts,omitempty"`
	Close bool           `json:"close,omitempty" msgpack:"close,omitempty"`
	Rooms []socket.Room  `json:"rooms,omitempty" msgpack:"rooms,omitempty"`
}
