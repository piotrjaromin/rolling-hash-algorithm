package sync

import (
	"bytes"

	// using gob for simplicity, but we could also write bytes directly (Takes less space)
	"encoding/gob"
	"io"
)

const byteBase = 16 * 16

func DeserializeChunks(chunksReader io.Reader) (map[uint32][]Chunk, error) {
	mappedChunks := map[uint32][]Chunk{}
	chunks := []Chunk{}

	enc := gob.NewDecoder(chunksReader)

	err := enc.Decode(&chunks)
	if err != nil {
		return mappedChunks, err
	}

	for _, chunk := range chunks {
		listOfChunks, ok := mappedChunks[chunk.RollingHash]
		if !ok {
			mappedChunks[chunk.RollingHash] = append(listOfChunks, chunk)
		} else {
			mappedChunks[chunk.RollingHash] = []Chunk{chunk}
		}
	}

	return mappedChunks, nil
}

func SerializeChunks(chunks []Chunk) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(chunks)
	if err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
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
