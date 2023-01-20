package sync

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Signature(t *testing.T) {

	tests := []struct {
		name     string
		data     []byte
		expected []Chunk
	}{
		{
			"calculates correct values for chunk smaller than defaultChunkSize",
			[]byte{1, 2, 3, 4, 5},
			[]Chunk{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := New()
			data := bytes.NewReader(test.data)

			chunks := []Chunk{}
			s.Signature(data, func(c Chunk) {
				chunks = append(chunks, c)
			})

			assert.Equal(t, len(chunks), 1)

			assert.Equal(t, chunks, []Chunk{
				{
					Id:          0,
					RollingHash: 501678095,
					StrongHash:  []byte{156, 207, 10, 110, 152, 18, 87, 240, 164, 1, 77, 214, 225, 229, 200, 10},
				},
			})
		})
	}
}

func prettyPrintBytes(d []byte) {
	for _, r := range d {
		fmt.Printf("%d, ", r)
	}
}
