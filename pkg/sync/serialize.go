package sync

import (
	"bytes"
	"io"
	"io/ioutil"
)

const byteBase = 16 * 16
const weakHashSizeInBytes = 4
const strongHashSizeInBytes = 16
const chunkSizeInBytes = weakHashSizeInBytes + strongHashSizeInBytes

func deserializeChunks(chunksReader io.Reader) ([]Chunk, error) {
	chunks := []Chunk{}

	chunksRaw, err := ioutil.ReadAll(chunksReader)
	if err != nil {
		return chunks, err
	}

	index := 0
	for i := 0; i < len(chunksRaw); i += chunkSizeInBytes {
		weakHashEnd := i + weakHashSizeInBytes
		rollingHashBytes := bytesToUint32(chunksRaw[i:weakHashEnd])
		strongHashBytes := chunksRaw[weakHashEnd : weakHashEnd+strongHashSizeInBytes]

		chunks = append(chunks, Chunk{
			Id:          index,
			StrongHash:  strongHashBytes,
			RollingHash: rollingHashBytes,
		})
		index++
	}

	return chunks, nil
}

func SerializeChunks(chunks []Chunk) []byte {
	b := new(bytes.Buffer)

	for _, chunk := range chunks {
		b.Write(uint32ToBytes(chunk.RollingHash))
		b.Write(chunk.StrongHash)
	}

	return b.Bytes()
}

func uint32ToBytes(val uint32) []byte {
	data := [4]byte{}

	for i := 3; i >= 0; i-- {
		res := val % byteBase
		val = val / byteBase
		data[i] = byte(res)
	}

	return data[:]
}

func bytesToUint32(vals []byte) uint32 {
	var result uint32 = 0
	var base uint32 = 1

	for i := 3; i >= 0; i-- {
		result += uint32(vals[i]) * base
		base *= byteBase
	}

	return result
}
