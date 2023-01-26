package sync

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	// MD4 is cryptographically broken, but we do it for hashes
	"github.com/piotrjaromin/rolling-hash-algorithm/pkg/rollinghash"
	_ "golang.org/x/crypto/md4"
)

// not efficient size but for simplicity
const defaultChunkSize = 16
const defaultBufferMultiplier = 3

type sync struct {
	chunkSizeInBytes int
	hasher           hash.Hash
	// instead of relaying on struct we should expect interface as rollingHash
	// so in future we could easily replace implementation
	rHash *rollinghash.RollingHash
}

type Chunk struct {
	Id          uint32
	RollingHash uint32
	StrongHash  []byte
}

type ChunkHandler func(Chunk)

type Operation byte

const (
	Init Operation = iota
	NewData
	ExistingData
)

type Delta struct {
	Id        uint32
	Operation Operation
	Data      []byte
}

type DeltaHandler func(Delta)

func New() sync {
	return sync{
		chunkSizeInBytes: defaultChunkSize,
		hasher:           crypto.MD4.New(),
		rHash:            rollinghash.New(uint32(defaultChunkSize)),
	}
}

func (r *sync) Signature(data io.Reader, handleChunks ChunkHandler) error {
	// we will read more bytes
	// but hashing will take into consideration only r.chunkSizeInBytes
	fullBufferSize := defaultBufferMultiplier * r.chunkSizeInBytes
	buffer := make([]byte, fullBufferSize)
	r.hasher.Reset()
	r.rHash.Reset()

	bytesLeft := 0
	var chunkIndex uint32 = 0
	for {
		n, err := data.Read(buffer[bytesLeft:])

		if n == 0 || err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// iterate until we run out of data
		i := 0
		for i <= n-r.chunkSizeInBytes {
			rollingChunk := buffer[i : i+r.chunkSizeInBytes]
			r.processChunk(chunkIndex, rollingChunk, handleChunks)
			chunkIndex++
			i += r.chunkSizeInBytes
		}

		// we need bytes from next chunk to process this
		bytesLeft = n - i
		for i := 0; i < bytesLeft; i++ {
			buffer[i] = buffer[n-bytesLeft+i]
		}

	}

	if bytesLeft > 0 {
		r.processChunk(chunkIndex, buffer[:bytesLeft], handleChunks)
	}
	return nil
}

func (r *sync) processChunk(chunkIndex uint32, rollingChunk []byte, handleChunks ChunkHandler) {
	rHash := r.rHash.AddBuffer(rollingChunk).Hash()

	r.hasher.Write(rollingChunk)
	handleChunks(Chunk{
		Id:          chunkIndex,
		RollingHash: rHash,
		StrongHash:  r.hasher.Sum(nil),
	})
	r.hasher.Reset()
}

func (r *sync) Delta(data io.Reader, chunksReader io.Reader, handleDeltas DeltaHandler) error {
	chunksList, err := DeserializeChunks(chunksReader)
	if err != nil {
		return fmt.Errorf("unable to deserialize signature file. %w", err)
	}
	chunks := chunksListToMap(chunksList)

	fullBufferSize := defaultBufferMultiplier * r.chunkSizeInBytes
	buffer := make([]byte, fullBufferSize)
	r.hasher.Reset()
	r.rHash.Reset()

	// we read in chunks, it maybe that we cannot proce
	bytesLeft := 0
	bytesProcessed := 0

	lastOperation := Init // Used to init rHash
	var deltaIndex uint32 = 0

	for {
		n, err := data.Read(buffer[bytesLeft:])
		if err != io.EOF && err != nil {
			return err
		}

		// n should be total of available bytes
		n = bytesLeft + n
		chunkSize := r.chunkSizeInBytes

		if n < r.chunkSizeInBytes {
			bytesProcessed = n
			chunkSize = n
		}

		i := 0
		for {
			if n < i+chunkSize || n == 0 {
				break
			}

			if lastOperation == Init {
				r.rHash.AddBuffer(buffer[0:chunkSize])
			} else {
				r.rHash.AddBuffer(buffer[i : i+bytesProcessed])
			}

			bytesProcessed, lastOperation = r.processBytesForDelta(
				lastOperation, buffer[i:i+chunkSize], chunks, deltaIndex, handleDeltas,
			)

			i += bytesProcessed
			deltaIndex += 1
		}

		bytesLeft = n - i

		// Two scenarios
		// 1. reading data, this is last iteration and there is leftover
		// 2. first iteration with chunk smaller than whole thing
		for j := 0; j < bytesLeft; j++ {
			buffer[j] = buffer[i+j]
		}

		if n == 0 || err == io.EOF {
			if bytesLeft == 0 {
				return nil
			}
		}
	}

}

func chunksListToMap(chunks []Chunk) map[uint32][]Chunk {
	mappedChunks := map[uint32][]Chunk{}

	for _, chunk := range chunks {
		listOfChunks, ok := mappedChunks[chunk.RollingHash]
		if !ok {
			mappedChunks[chunk.RollingHash] = append(listOfChunks, chunk)
		} else {
			mappedChunks[chunk.RollingHash] = []Chunk{chunk}
		}
	}

	return mappedChunks
}

// TODO better name....
func (r *sync) processBytesForDelta(
	lastOperation Operation, buffer []byte,
	chunks map[uint32][]Chunk, deltaIndex uint32, handleDeltas DeltaHandler,
) (int, Operation) {

	foundStrongHashMatch := false
	fromChunks, ok := chunks[r.rHash.Hash()]
	if ok {
		r.hasher.Write(buffer)
		strongHash := r.hasher.Sum(nil)
		foundStrongHashMatch = handleStrongHashCheck(fromChunks, strongHash, deltaIndex, handleDeltas)
		r.hasher.Reset()
	}

	if foundStrongHashMatch {
		return len(buffer), ExistingData
	} else {
		// if no match then send new bytes to be added to file
		handleDeltas(Delta{
			Id:        deltaIndex,
			Operation: NewData,
			Data:      []byte{buffer[0]},
		})

		return 1, NewData
	}
}

func handleStrongHashCheck(fromChunks []Chunk, strongHash []byte, deltaIndex uint32, handleDeltas DeltaHandler) bool {
	// iterate over list and check for strong hash
	for _, chunk := range fromChunks {

		// if strong hash match then send that original file contains data
		if bytes.Equal(strongHash, chunk.StrongHash) {
			handleDeltas(Delta{
				Id:        deltaIndex,
				Operation: ExistingData,
				Data:      uint32ToBytes(chunk.Id),
			})
			return true
		}
	}

	return false
}

func toHex(d []byte) string {
	str := ""

	for _, val := range d {
		str += fmt.Sprintf("0x%s, ", hex.EncodeToString([]byte{val}))
	}

	return str
}
