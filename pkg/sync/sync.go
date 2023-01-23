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
			r.processChunk(rollingChunk, handleChunks)
		}

		bytesLeft := n % r.chunkSizeInBytes
		if bytesLeft > 0 {
			rollingChunk := buffer[n-bytesLeft:]
			r.processChunk(rollingChunk, handleChunks)
		}
	}
}

func (r *sync) processChunk(rollingChunk []byte, handleChunks ChunkHandler) {
	rHash := rollinghash.InitRollingHash(rollingChunk).Hash()

	r.hasher.Write(rollingChunk)
	handleChunks(Chunk{
		RollingHash: rHash,
		StrongHash:  r.hasher.Sum(nil),
	})
	r.hasher.Reset()
}

func (r *sync) Delta(data io.Reader, chunksReader io.Reader, handleDeltas DeltaHandler) error {
	buffer := make([]byte, defaultBufferMultiplier*r.chunkSizeInBytes)

	chunks, err := DeserializeChunks(chunksReader)
	if err != nil {
		return fmt.Errorf("unable to deserialize signature file. %w", err)
	}

	r.hasher.Reset()

	// instead of relaying on struct we should expect interface as rollingHash
	// so in futre we could easily replace implementation
	var rHash *rollinghash.RollingHash = nil
	// we read in chunks, it maybe that we cannot proce
	bytesLeft := 0

	for {
		n, err := data.Read(buffer[bytesLeft:])

		if n == 0 || err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if rHash == nil {
			initChunk := buffer[:r.chunkSizeInBytes]
			rHash = rollinghash.InitRollingHash(initChunk)
		}

		for i := 0; i < n-r.chunkSizeInBytes; i += 1 {
			fromChunks, ok := chunks[rHash.Hash()]
			if ok {

				r.hasher.Write(buffer[i:r.chunkSizeInBytes])
				strongHash := r.hasher.Sum(nil)
				// iterate over list and check for strong hash
				for _, chunk := range fromChunks {

					// if strong hash match then send operation
					if bytes.Equal(strongHash, chunk.StrongHash) {
						handleDeltas(Delta{
							Operation: ExistingData,
							Data:      uint32ToBytes(chunk.Id),
						})
					}

					i += r.chunkSizeInBytes
					break
				}
			}

			// if no match then send existing bits
			handleDeltas(Delta{
				Operation: NewData,
				Data:      buffer[i:r.chunkSizeInBytes],
			})

			rHash.Add(buffer[i+1])
		}

		// we need bytes from next chunk to process this
		bytesLeft = n % r.chunkSizeInBytes
		for i := 0; i < bytesLeft; i++ {
			buffer[i] = buffer[n-bytesLeft+i]
		}
	}
}
