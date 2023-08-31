package emitter

import (
	"github.com/zishang520/socket.io/socket"
)

const UID = "emitter"

type RequestType int

const (
	SOCKETS           RequestType = 0x0
	ALL_ROOMS         RequestType = 0x1
	REMOTE_JOIN       RequestType = 0x2
	REMOTE_LEAVE      RequestType = 0x3
	REMOTE_DISCONNECT RequestType = 0x4
	REMOTE_FETCH      RequestType = 0x5
	SERVER_SIDE_EMIT  RequestType = 0x6
)

type Parser interface {
	Encode(any) ([]byte, error)
}

type BroadcastOptions struct {
	Nsp              string
	BroadcastChannel string
	RequestChannel   string
	Parser           Parser
}

type PacketOptions struct {
	Rooms  []socket.Room          `json:"rooms,omitempty" mapstructure:"rooms,omitempty" msgpack:"rooms,omitempty"`
	Except []socket.Room          `json:"except,omitempty" mapstructure:"except,omitempty" msgpack:"except,omitempty"`
	Flags  *socket.BroadcastFlags `json:"flags,omitempty" mapstructure:"except,omitempty" msgpack:"flags,omitempty"`
}

type Packet struct {
	Uid   string         `json:"uid,omitempty" msgpack:"uid,omitempty"`
	Type  RequestType    `json:"type,omitempty" msgpack:"type,omitempty"`
	Data  any            `json:"data,omitempty" msgpack:"data,omitempty"`
	Opts  *PacketOptions `json:"opts,omitempty" msgpack:"opts,omitempty"`
	Close bool           `json:"close,omitempty" msgpack:"close,omitempty"`
	Rooms []socket.Room  `json:"rooms,omitempty" msgpack:"rooms,omitempty"`
}
