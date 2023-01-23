package rollinghash

const moduloVal uint32 = 1 << 16

type RollingHash struct {
	buffer      []byte
	a           uint32
	b           uint32
	l           uint32
	initialized bool
}

func New(bufferSize uint32) *RollingHash {
	return &RollingHash{
		initialized: false,
		buffer:      make([]byte, bufferSize),
		l:           bufferSize,
	}
}

// TODO not needed??
// func initRollingHash(data []byte) *RollingHash {
// 	l := len(data)

// 	var a uint32 = 0
// 	var b uint32 = 0

// 	for i, val := range data {
// 		a += uint32(val)
// 		b += uint32(l-i) * uint32(val)
// 	}

// 	a = a % moduloVal
// 	b = b % moduloVal

// 	buffer := make([]byte, len(data))
// 	copy(buffer, data)

// 	return &RollingHash{
// 		buffer: buffer,
// 		a:      uint32(a),
// 		b:      uint32(b),
// 		l:      uint32(l),
// 	}
// }

func (r *RollingHash) Add(b byte) *RollingHash {
	r.a = (r.a - uint32(r.buffer[0]) + uint32(b)) % moduloVal
	r.b = (r.b - (r.l)*uint32(r.buffer[0]) + r.a) % moduloVal

	for i := 0; i < int(r.l)-1; i++ {
		r.buffer[i] = r.buffer[i+1]
	}

	r.buffer[r.l-1] = b
	return r
}

func (r *RollingHash) AddBuffer(data []byte) *RollingHash {
	for _, b := range data {
		r.Add(b)
	}
	return r
}

func (r *RollingHash) Hash() uint32 {
	s := r.a + moduloVal*r.b
	return s
}
