package sync

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Signature(t *testing.T) {

	tests := []struct {
		name     string
		data     func() []byte
		expected []Chunk
	}{
		{
			"calculates correct values for chunk smaller than defaultChunkSize",
			func() []byte {
				return []byte{1, 2, 3, 4, 5}
			},
			[]Chunk{
				{
					Id:          0,
					RollingHash: 2293775,
					StrongHash:  []byte{147, 235, 175, 223, 237, 209, 153, 78, 128, 24, 204, 41, 92, 193, 168, 238},
				},
			},
		},
		{
			"calculates correct values for input which is multiplication of defaultChunkSize",
			func() []byte {
				data, _ := dataGenerateRandom(defaultChunkSize * 3)
				return data
			},
			[]Chunk{
				{
					Id:          0,
					RollingHash: 1042155156,
					StrongHash:  []byte{94, 60, 176, 51, 111, 119, 24, 198, 245, 90, 183, 135, 67, 151, 254, 92},
				},
				{
					Id:          1,
					RollingHash: 1350305989,
					StrongHash:  []byte{131, 178, 254, 182, 230, 165, 19, 207, 96, 156, 123, 23, 212, 232, 60, 142},
				},
				{
					Id:          2,
					RollingHash: 1137182653,
					StrongHash:  []byte{97, 22, 8, 89, 223, 73, 192, 55, 73, 2, 199, 154, 68, 152, 240, 42},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := New()
			data := bytes.NewReader(test.data())

			chunks := []Chunk{}
			s.Signature(data, func(c Chunk) {
				chunks = append(chunks, c)
			})

			assert.Equal(t, test.expected, chunks)
		})
	}
}

func Test_DeltaReturnsNoChangesWhenNewFileIsTheSameAsOld(t *testing.T) {
	dataSize := 20002
	data, dataReader := dataGenerateRandom(dataSize)

	chunks := []Chunk{}

	s := New()
	s.Signature(dataReader, func(c Chunk) {
		chunks = append(chunks, c)
	})

	chunksAsBytes, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	var currentOperationId uint32
	s.Delta(bytes.NewReader(data), chunksAsBytes, func(d Delta) {
		assert.Equal(t, ExistingData, d.Operation, "Invalid operation for operation id: %d", currentOperationId)
		require.Equal(t, currentOperationId, d.Id, "Mismatch with expected operation id")
		currentOperationId += 1
	})

	expectedLastOperationId := uint32(math.Ceil(float64(dataSize) / defaultChunkSize))
	assert.Equal(t, expectedLastOperationId, currentOperationId, "expected %d to be %d", currentOperationId, expectedLastOperationId)
}

func Test_PassesForFilesSmallerThankChunkSize(t *testing.T) {
	dataSize := defaultChunkSize - 1
	_, dataReader := dataGenerateRandom(dataSize)

	chunks := []Chunk{}

	s := New()
	s.Signature(dataReader, func(c Chunk) {
		chunks = append(chunks, c)
	})

	require.Len(t, chunks, 1)

	_, sameDataReader := dataGenerateRandom(dataSize)

	chunksAsBytes, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	var expectedOperationId uint32
	s.Delta(sameDataReader, chunksAsBytes, func(d Delta) {
		assert.Equal(t, ExistingData, d.Operation, "Invalid operation for operation id: %d", expectedOperationId)
		require.Equal(t, expectedOperationId, d.Id, "Mismatch with expected operation id")
		require.Equal(t, chunks[0].Id, bytesToUint32(d.Data))
		expectedOperationId += 1
	})

	require.Equal(t, uint32(1), expectedOperationId)
}

func Test_SendsNewDataForCompletelyNewFile(t *testing.T) {
	dataSize := 50
	_, dataReader := dataGenerateRandom(dataSize)

	chunks := []Chunk{}

	s := New()
	s.Signature(dataReader, func(c Chunk) {
		chunks = append(chunks, c)
	})

	chunksAsBytes, err := SerializeChunks(chunks)
	assert.Nil(t, err)

	newFileSize := 80
	newFileBytes, newFile := dataGenerateRandomWithSeed(newFileSize, 500)

	var expectedOperationId uint32
	receivedBytes := []byte{}
	s.Delta(newFile, chunksAsBytes, func(d Delta) {
		assert.Equal(t, NewData, d.Operation, "Invalid operation for operation id: %d", expectedOperationId)
		require.Equal(t, expectedOperationId, d.Id, "Mismatch with expected operation id")
		expectedOperationId += 1
		receivedBytes = append(receivedBytes, d.Data...)
	})

	require.Equal(t, len(newFileBytes), len(receivedBytes))
	require.Equal(t, newFileBytes, receivedBytes)
	require.Equal(t, uint32(newFileSize), expectedOperationId)
}

func Test_DeltaInformsThatFileWasPrependedWithNewData(t *testing.T) {

}

func Test_DeltaInformsThatFileWasSuffixedWithNewData(t *testing.T) {

}

func Test_DeltaInformsThatFileWasInsertedWithNewData(t *testing.T) {

}

func dataGenerateRandom(size int) ([]byte, io.Reader) {
	return dataGenerateRandomWithSeed(size, 20)
}

func dataGenerateRandomWithSeed(size int, seed int64) ([]byte, io.Reader) {
	buffer := make([]byte, size)
	rand.Seed(seed)

	for i := range buffer {
		buffer[i] = byte(rand.Int())
	}

	return buffer, bytes.NewReader(buffer)
}
