package sync

import (
	"crypto"
	"hash"
	"io"

	// MD4 is cryptographically broken, but we do it for hashes
	_ "golang.org/x/crypto/md4"
)

const defaultChunkSize = 16
const defaultBufferMultiplier = 32

type sync struct {
	chunkSizeInBytes int
	hasher           hash.Hash
}

type Chunk struct {
	Id          uint64
	RollingHash uint32
	StrongHash  []byte
}

type ChunkHandler func(Chunk)

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
	rollingHash := InitRollingHash(rollingChunk).Hash()

	r.hasher.Write(rollingChunk)
	handleChunks(Chunk{
		RollingHash: rollingHash,
		StrongHash:  r.hasher.Sum(nil),
	})
	r.hasher.Reset()
}
