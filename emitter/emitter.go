package emitter

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/zishang520/socket.io/socket"
)

type Emitter struct {
	redisClient      *redis.Client
	opts             *EmitterOptions
	broadcastOptions *BroadcastOptions
	ctx              context.Context
}

func NewEmitter(redisClient *redis.Client, opts *EmitterOptions, nsps ...string) *Emitter {
	if redisClient == nil {
		panic("redisClient must not be nil")
	}

	e := &Emitter{}
	e.ctx = context.Background()
	e.redisClient = redisClient

	if opts == nil {
		opts = DefaultEmitterOptions()
	}
	e.opts = opts

	nsp := "/"
	if len(nsps) > 0 {
		nsp = nsps[0]
		if len(nsp) == 0 {
			nsp = "/"
		}
	}

	e.broadcastOptions = &BroadcastOptions{
		Nsp:              nsp,
		BroadcastChannel: e.opts.Key() + "#" + nsp + "#",
		RequestChannel:   e.opts.Key() + "-request#" + nsp + "#",
		Parser:           e.opts.Parser(),
	}
	return e
}

// Return a new emitter for the given namespace.
func (e *Emitter) Of(nsp string) *Emitter {
	if len(nsp) > 0 {
		if nsp[0] != '/' {
			nsp = "/" + nsp
		}
	} else {
		nsp = "/"
	}
	return NewEmitter(e.redisClient, e.opts, nsp)
}

// Emits to all clients.
func (e *Emitter) Emit(ev string, args ...any) error {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).Emit(ev, args...)
}

// Targets a room when emitting.
func (e *Emitter) To(room ...socket.Room) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).To(room...)
}

// Targets a room when emitting.
func (e *Emitter) In(room ...socket.Room) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).In(room...)
}

// Excludes a room when emitting.
func (e *Emitter) Except(room ...socket.Room) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).Except(room...)
}

// Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
// receive messages (because of network slowness or other issues, or because they’re connected through long polling
// and is in the middle of a request-response cycle).
func (e *Emitter) Volatile() *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).Volatile()
}

// Sets the compress flag.
//
// compress - if `true`, compresses the sending data
func (e *Emitter) Compress(compress bool) *BroadcastOperator {
	return NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).Compress(compress)
}

// Makes the matching socket instances join the specified rooms
func (e *Emitter) SocketsJoin(room ...socket.Room) {
	NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).SocketsJoin(room...)
}

// Makes the matching socket instances leave the specified rooms
func (e *Emitter) SocketsLeave(room ...socket.Room) {
	NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).SocketsLeave(room...)
}

// Makes the matching socket instances disconnect
func (e *Emitter) DisconnectSockets(state bool) {
	NewBroadcastOperator(e.redisClient, e.broadcastOptions, nil, nil, nil).DisconnectSockets(state)
}

// Send a packet to the Socket.IO servers in the cluster
//
// args - any number of serializable arguments
func (e *Emitter) ServerSideEmit(args ...any) error {
	if _, withAck := args[len(args)-1].(func(...any)); withAck {
		return errors.New("Acknowledgements are not supported")
	}
	request, err := json.Marshal(&Packet{
		Uid:  UID,
		Type: SERVER_SIDE_EMIT,
		Data: args,
	})
	if err != nil {
		return err
	}
	return e.redisClient.Publish(e.ctx, e.broadcastOptions.RequestChannel, request).Err()
}
