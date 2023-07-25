package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"
)

type Msgpack struct {
}

func New() *Msgpack {
	return &Msgpack{}
}

func (m *Msgpack) Encoder(value any) ([]byte, error) {
	return msgpack.Marshal(value)
}

func (m *Msgpack) Decoder(dara []byte, value any) error {
	return msgpack.Unmarshal(dara, value)
}
