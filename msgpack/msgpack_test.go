package msgpack

import (
	"bytes"
	"testing"
)

func TestEmitterOptions(t *testing.T) {
	pack := New()

	t.Run("Encoder/Decoder", func(t *testing.T) {
		data, err := pack.Encoder([]any{[]byte{1, 2, 3, 4}, 0, 4, nil})
		if err != nil {
			t.Fatal("Encoder error must be nil")
		}
		check := []byte{148, 196, 4, 1, 2, 3, 4, 0, 4, 192}
		if !bytes.Equal(data, check) {
			t.Fatalf(`Encoder value not as expected: %v, want match for %v`, data, check)
		}
		var value any
		err = pack.Decoder(data, &value)
		if err != nil {
			t.Fatal("Decoder error must be nil")
		}
		if d, ok := value.([]any); !ok {
			t.Fatal("Decoder value must be []any")
		} else {
			if n := len(d); n != 4 {
				t.Fatalf(`Decoder len(value) not as expected: %v, want match for %v`, 4, n)
			}
			if d[3] != nil {
				t.Fatalf(`Decoder value[3] not as expected: %v, want match for %v`, nil, d[3])
			}
		}
	})
}
