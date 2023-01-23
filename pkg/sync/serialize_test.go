package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Uint32ToBytesShouldConvertCorrectly(t *testing.T) {
	var val uint32 = 12343
	bytes := uint32ToBytes(val)

	assert.Equal(t, []byte{0, 0, 48, 55}, bytes)
}

func Test_BytesToUint32ShouldConvertCorrectly(t *testing.T) {
	bytes := []byte{3, 59, 255, 10}

	val := bytesToUint32(bytes)

	assert.Equal(t, uint32(54263562), val)
}

func Test_ByteAndUint32ConversionShouldWorkBothWays(t *testing.T) {
	var val uint32 = 435432

	bytes := uint32ToBytes(val)
	newVal := bytesToUint32(bytes)

	assert.Equal(t, val, newVal)
}

func Test_ChunksSerializationShouldWorkBothWays(t *testing.T) {
	chunks := []Chunk{
		{
			0,
			83712,
			[]byte{156, 207, 10, 110, 152, 18, 87, 240, 164, 1, 77, 214, 225, 229, 200, 10},
		},
		{
			1,
			12343,
			[]byte{86, 38, 10, 24, 122, 218, 87, 43, 164, 4, 77, 214, 225, 229, 203, 55},
		},
	}

	reader, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	readChunks, err := DeserializeChunks(reader)

	assert.Nil(t, err)

	assert.Equal(t, chunks, readChunks)
}
