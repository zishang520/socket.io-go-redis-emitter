package emitter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io-go-parser/parser"
	"github.com/zishang520/socket.io/socket"
)

var emitter_log = log.NewLog("socket.io-emitter")

var RESERVED_EVENTS = types.NewSet(
	"connect",
	"connect_error",
	"disconnect",
	"disconnecting",
	"newListener",
	"removeListener",
)

type BroadcastOperator struct {
	redisClient      *redis.Client
	broadcastOptions *BroadcastOptions
	rooms            *types.Set[socket.Room]
	exceptRooms      *types.Set[socket.Room]
	flags            *socket.BroadcastFlags
	ctx              context.Context
}

func NewBroadcastOperator(redisClient *redis.Client, broadcastOptions *BroadcastOptions, rooms *types.Set[socket.Room], exceptRooms *types.Set[socket.Room], flags *socket.BroadcastFlags) *BroadcastOperator {
	if redisClient == nil {
		panic("redisClient must not be nil")
	}

	if broadcastOptions == nil {
		panic("broadcastOptions must not be nil")
	}

	b := &BroadcastOperator{}
	b.ctx = context.Background()
	b.redisClient = redisClient
	b.broadcastOptions = broadcastOptions

	if rooms == nil {
		b.rooms = types.NewSet[socket.Room]()
	} else {
		b.rooms = rooms
	}

	if exceptRooms == nil {
		b.exceptRooms = types.NewSet[socket.Room]()
	} else {
		b.exceptRooms = exceptRooms
	}

	if flags == nil {
		b.flags = &socket.BroadcastFlags{}
	} else {
		b.flags = flags
	}

	return b
}

// Targets a room when emitting.
func (b *BroadcastOperator) To(room ...socket.Room) *BroadcastOperator {
	rooms := types.NewSet(b.rooms.Keys()...)
	rooms.Add(room...)
	return NewBroadcastOperator(b.redisClient, b.broadcastOptions, rooms, b.exceptRooms, b.flags)
}

// Targets a room when emitting.
func (b *BroadcastOperator) In(room ...socket.Room) *BroadcastOperator {
	return b.To(room...)
}

// Excludes a room when emitting.
func (b *BroadcastOperator) Except(room ...socket.Room) *BroadcastOperator {
	exceptRooms := types.NewSet(b.exceptRooms.Keys()...)
	exceptRooms.Add(room...)
	return NewBroadcastOperator(b.redisClient, b.broadcastOptions, b.rooms, exceptRooms, b.flags)
}

// Sets the compress flag.
func (b *BroadcastOperator) Compress(compress bool) *BroadcastOperator {
	flags := *b.flags
	flags.Compress = compress
	return NewBroadcastOperator(b.redisClient, b.broadcastOptions, b.rooms, b.exceptRooms, &flags)
}

// Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
// receive messages (because of network slowness or other issues, or because they’re connected through long polling
// and is in the middle of a request-response cycle).
func (b *BroadcastOperator) Volatile() *BroadcastOperator {
	flags := *b.flags
	flags.Volatile = true
	return NewBroadcastOperator(b.redisClient, b.broadcastOptions, b.rooms, b.exceptRooms, &flags)
}

// Emits to all clients.
func (b *BroadcastOperator) Emit(ev string, args ...any) error {
	if RESERVED_EVENTS.Has(ev) {
		return errors.New(fmt.Sprintf(`"%s" is a reserved event name`, ev))
	}

	if b.broadcastOptions.Parser == nil {
		return errors.New(`broadcastOptions.Parser is not set`)
	}

	// set up packet object
	data := append([]any{ev}, args...)

	packet := &parser.Packet{
		Type: parser.EVENT,
		Nsp:  b.broadcastOptions.Nsp,
		Data: data,
	}

	opts := &PacketOptions{
		Rooms:  b.rooms.Keys(),
		Except: b.exceptRooms.Keys(),
		Flags:  b.flags,
	}

	msg, err := b.broadcastOptions.Parser.Encode([]any{UID, packet, opts})
	if err != nil {
		return nil
	}

	channel := b.broadcastOptions.BroadcastChannel
	if b.rooms != nil && b.rooms.Len() == 1 {
		channel += string((b.rooms.Keys())[0]) + "#"
	}

	emitter_log.Debug("publishing message to channel %s", channel)

	return b.redisClient.Publish(b.ctx, channel, msg).Err()
}

// Makes the matching socket instances join the specified rooms
func (b *BroadcastOperator) SocketsJoin(room ...socket.Room) {
	request, err := json.Marshal(&Packet{
		Type: REQUEST_REMOTE_JOIN,
		Opts: &PacketOptions{
			Rooms:  b.rooms.Keys(),
			Except: b.exceptRooms.Keys(),
		},
		Rooms: room,
	})
	if err != nil {
		return
	}
	b.redisClient.Publish(b.ctx, b.broadcastOptions.RequestChannel, request)
}

// Makes the matching socket instances leave the specified rooms
func (b *BroadcastOperator) SocketsLeave(room ...socket.Room) {
	request, err := json.Marshal(&Packet{
		Type: REQUEST_REMOTE_LEAVE,
		Opts: &PacketOptions{
			Rooms:  b.rooms.Keys(),
			Except: b.exceptRooms.Keys(),
		},
		Rooms: room,
	})
	if err != nil {
		return
	}
	b.redisClient.Publish(b.ctx, b.broadcastOptions.RequestChannel, request)
}

// Makes the matching socket instances disconnect
func (b *BroadcastOperator) DisconnectSockets(state bool) {
	request, err := json.Marshal(&Packet{
		Type: REQUEST_REMOTE_DISCONNECT,
		Opts: &PacketOptions{
			Rooms:  b.rooms.Keys(),
			Except: b.exceptRooms.Keys(),
		},
		Close: state,
	})
	if err != nil {
		return
	}
	b.redisClient.Publish(b.ctx, b.broadcastOptions.RequestChannel, request)
}
