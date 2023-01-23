package sync

import (
	"bytes"
	"io"
	"math/rand"
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

			expectedChunks := []Chunk{
				{
					Id:          0,
					RollingHash: 500695055,
					StrongHash:  []byte{156, 207, 10, 110, 152, 18, 87, 240, 164, 1, 77, 214, 225, 229, 200, 10},
				},
			}

			assert.Equal(t, expectedChunks, chunks)
		})
	}
}

func Test_DeltaReturnsNoChangesWhenNewFileIsTheSameAsOld(t *testing.T) {
	dataSize := 50
	data, dataReader := dataGenerateRandom(dataSize)

	chunks := []Chunk{}

	s := New()
	s.Signature(dataReader, func(c Chunk) {
		chunks = append(chunks, c)
	})

	chunksAsBytes, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	expectedOperationId := 0
	s.Delta(bytes.NewReader(data), chunksAsBytes, func(d Delta) {
		assert.Equal(t, ExistingData, d.Operation)
		assert.Equal(t, expectedOperationId, d.Id)
		expectedOperationId += 1
	})

	assert.Equal(t, 1250, expectedOperationId)
}

func Test_PassesForFilesSmallerThankChunkSize(t *testing.T) {
	dataSize := defaultChunkSize - 1
	_, dataReader := dataGenerateRandom(dataSize)

	chunks := []Chunk{}

	s := New()
	s.Signature(dataReader, func(c Chunk) {
		chunks = append(chunks, c)
	})

	_, sameDataReader := dataGenerateRandom(dataSize)

	chunksAsBytes, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	expectedOperationId := 0
	s.Delta(sameDataReader, chunksAsBytes, func(d Delta) {
		assert.Equal(t, ExistingData, d.Operation)
		assert.Equal(t, expectedOperationId, d.Id)
		expectedOperationId += 1
	})

	assert.Equal(t, 1, expectedOperationId)
}

func Test_DeltaInformsThatFileWasPrependedWithNewData(t *testing.T) {

}

func Test_DeltaInformsThatFileWasSuffixedWithNewData(t *testing.T) {

}

func Test_DeltaInformsThatFileWasInsertedWithNewData(t *testing.T) {

}

func dataGenerate(size int, val byte) ([]byte, io.Reader) {
	d := make([]byte, size)

	for i := range d {
		d[i] = val
	}

	return d, bytes.NewReader(d)
}

func dataGenerateRandom(size int) ([]byte, io.Reader) {
	d := make([]byte, size)
	rand.Seed(20)

	for i := range d {
		d[i] = byte(rand.Int())
	}

	return d, bytes.NewReader(d)
}
