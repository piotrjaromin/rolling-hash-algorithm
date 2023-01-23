package sync

import (
	"bytes"
	"crypto"
	"fmt"
	"hash"
	"io"

	// MD4 is cryptographically broken, but we do it for hashes
	"github.com/piotrjaromin/rolling-hash-algorithm/pkg/rollinghash"
	_ "golang.org/x/crypto/md4"
)

// not efficient size but for simplicity
const defaultChunkSize = 16
const defaultBufferMultiplier = 32

type sync struct {
	chunkSizeInBytes int
	hasher           hash.Hash
}

type Chunk struct {
	Id          uint32
	RollingHash uint32
	StrongHash  []byte
}

type ChunkHandler func(Chunk)

type Operation byte

const (
	Nop Operation = iota
	NewData
	ExistingData
)

type Delta struct {
	Id        int
	Operation Operation
	Data      []byte
}

type DeltaHandler func(Delta)

func New() sync {
	return sync{
		chunkSizeInBytes: defaultChunkSize,
		hasher:           crypto.MD4.New(),
	}
}

func (r *sync) Signature(data io.Reader, handleChunks ChunkHandler) error {
	// we will read more bytes
	// but hashing will take into consideration only r.chunkSizeInBytes
	buffer := make([]byte, defaultBufferMultiplier*r.chunkSizeInBytes)
	r.hasher.Reset()

	var chunkIndex uint32 = 0
	for {
		n, err := data.Read(buffer)

		if n == 0 || err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		// iterate until we run out of data
		for i := 0; i < n-r.chunkSizeInBytes; i += r.chunkSizeInBytes {
			rollingChunk := buffer[i : i+r.chunkSizeInBytes]
			r.processChunk(chunkIndex, rollingChunk, handleChunks)
			chunkIndex++
		}

		bytesLeft := n % r.chunkSizeInBytes
		if bytesLeft > 0 {
			rollingChunk := buffer[n : n+bytesLeft]
			r.processChunk(chunkIndex, rollingChunk, handleChunks)
			chunkIndex++
		}
	}
}

func (r *sync) processChunk(chunkIndex uint32, rollingChunk []byte, handleChunks ChunkHandler) {
	rHash := rollinghash.InitRollingHash(rollingChunk).Hash()

	r.hasher.Write(rollingChunk)
	handleChunks(Chunk{
		Id:          chunkIndex,
		RollingHash: rHash,
		StrongHash:  r.hasher.Sum(nil),
	})
	r.hasher.Reset()
}

func (r *sync) Delta(data io.Reader, chunksReader io.Reader, handleDeltas DeltaHandler) error {
	buffer := make([]byte, defaultBufferMultiplier*r.chunkSizeInBytes)

	chunksList, err := DeserializeChunks(chunksReader)
	if err != nil {
		return fmt.Errorf("unable to deserialize signature file. %w", err)
	}

	chunks := chunksListToMap(chunksList)
	r.hasher.Reset()

	// instead of relaying on struct we should expect interface as rollingHash
	// so in future we could easily replace implementation
	var rHash *rollinghash.RollingHash = nil
	// we read in chunks, it maybe that we cannot proce
	bytesLeft := 0

	deltaIndex := 0
	for {
		n, err := data.Read(buffer[bytesLeft:])

		if n == 0 || err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if rHash == nil {
			// TODO if buffer is too small this will fail
			initChunk := buffer[:r.chunkSizeInBytes]
			rHash = rollinghash.InitRollingHash(initChunk)
		}

		i := 0
		for i < n-r.chunkSizeInBytes {
			foundStrongHashMatch := false
			fromChunks, ok := chunks[rHash.Hash()]
			if ok {
				r.hasher.Write(buffer[i : i+r.chunkSizeInBytes])
				strongHash := r.hasher.Sum(nil)
				foundStrongHashMatch = handleStrongHashCheck(fromChunks, strongHash, deltaIndex, handleDeltas)
				r.hasher.Reset()
			}

			if foundStrongHashMatch {
				rHash.AddBuffer(buffer[i : i+r.chunkSizeInBytes])
				i += r.chunkSizeInBytes
			} else {
				// if no match then send new bytes to be added to file
				handleDeltas(Delta{
					Operation: NewData,
					Data:      []byte{buffer[i]},
				})

				i += 1
				rHash.Add(buffer[i])
			}

			deltaIndex += 1
		}

		// we need bytes from next chunk to process this
		bytesLeft = n % r.chunkSizeInBytes
		for i := 0; i < bytesLeft; i++ {
			buffer[i] = buffer[n-bytesLeft+i]
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

func handleStrongHashCheck(fromChunks []Chunk, strongHash []byte, deltaIndex int, handleDeltas DeltaHandler) bool {
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
