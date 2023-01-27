package sync

import (
	"bytes"

	// using gob for simplicity, but we could also write bytes directly (Takes less space) or use protobuf
	"encoding/gob"
	"io"
)

const byteBase = 16 * 16

func DeserializeChunks(chunksReader io.Reader) ([]Chunk, error) {
	chunks := []Chunk{}

	enc := gob.NewDecoder(chunksReader)
	err := enc.Decode(&chunks)
	if err != nil {
		return chunks, err
	}

	return chunks, nil
}

func SerializeChunks(chunks []Chunk) (io.Reader, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(chunks)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

// could be generics
func DeserializeDelta(deltasReader io.Reader) ([]Delta, error) {
	deltas := []Delta{}

	enc := gob.NewDecoder(deltasReader)
	err := enc.Decode(&deltas)
	if err != nil {
		return deltas, err
	}

	return deltas, nil
}

func SerializeDeltas(deltas []Delta) (io.Reader, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(deltas)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
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

	for i := len(vals) - 1; i >= 0; i-- {
		result += uint32(vals[i]) * base
		base *= byteBase
	}

	return result
}
