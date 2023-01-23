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
	Id        int
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
		i := 0
		for i < n-r.chunkSizeInBytes {
			rollingChunk := buffer[i : i+r.chunkSizeInBytes]
			r.processChunk(chunkIndex, rollingChunk, handleChunks)
			chunkIndex++
			i += r.chunkSizeInBytes
		}

		bytesLeft := n - i
		if bytesLeft > 0 {
			rollingChunk := buffer[i : i+bytesLeft]
			r.processChunk(chunkIndex, rollingChunk, handleChunks)
			chunkIndex++
		}
	}
}

func (r *sync) processChunk(chunkIndex uint32, rollingChunk []byte, handleChunks ChunkHandler) {
	rHash := rollinghash.New(uint32(r.chunkSizeInBytes)).AddBuffer(rollingChunk).Hash()

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

	// we read in chunks, it maybe that we cannot proce
	bytesLeft := 0

	lastOperation := Init // Used to init rHash
	deltaIndex := 0
	for {
		n, err := data.Read(buffer[bytesLeft:])

		if n == 0 || err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		i := 0
		for i < n-r.chunkSizeInBytes {
			i += r.processBytes(
				lastOperation, i, buffer, chunks, deltaIndex, handleDeltas,
			)

			deltaIndex += 1
		}

		// we need bytes from next chunk to process this
		bytesLeft = n - i
		for i := 0; i < bytesLeft; i++ {
			buffer[i] = buffer[n-bytesLeft+i]
		}
	}

	r.processBytes(
		lastOperation, 0, buffer[:bytesLeft], chunks, deltaIndex, handleDeltas,
	)

	return nil
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
func (r *sync) processBytes(
	lastOperation Operation, i int, buffer []byte,
	chunks map[uint32][]Chunk, deltaIndex int, handleDeltas DeltaHandler,
) int {
	chunkSize := r.chunkSizeInBytes
	if len(buffer) < r.chunkSizeInBytes {
		chunkSize = len(buffer) //
	}

	foundStrongHashMatch := false

	if lastOperation == ExistingData || lastOperation == Init {
		r.rHash.AddBuffer(buffer[i : i+chunkSize])
	} else {
		r.rHash.Add(buffer[i])
	}

	fromChunks, ok := chunks[r.rHash.Hash()]
	if ok {
		r.hasher.Write(buffer[i : i+chunkSize])
		strongHash := r.hasher.Sum(nil)
		foundStrongHashMatch = handleStrongHashCheck(fromChunks, strongHash, deltaIndex, handleDeltas)
		r.hasher.Reset()
	}

	if foundStrongHashMatch {
		return chunkSize
	} else {
		// if no match then send new bytes to be added to file
		handleDeltas(Delta{
			Operation: NewData,
			Data:      []byte{buffer[i]},
		})

		return 1
	}
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
