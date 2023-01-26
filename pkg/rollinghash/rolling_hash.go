package rollinghash

const moduloVal uint32 = 1 << 16

type RollingHash struct {
	buffer             []byte
	addOperationsCount int
	a                  uint32
	b                  uint32
	l                  uint32
}

func New(bufferSize uint32) *RollingHash {
	return &RollingHash{
		buffer: make([]byte, bufferSize),
		l:      bufferSize,
	}
}

func (r *RollingHash) Add(b byte) *RollingHash {
	r.a = (r.a - uint32(r.buffer[0]) + uint32(b)) % moduloVal
	r.b = (r.b - (r.l)*uint32(r.buffer[0]) + r.a) % moduloVal

	for i := 0; i < int(r.l)-1; i++ {
		r.buffer[i] = r.buffer[i+1]
	}

	r.buffer[r.l-1] = b
	r.addOperationsCount += 1
	return r
}

func (r *RollingHash) AddBuffer(data []byte) *RollingHash {
	for _, b := range data {
		r.Add(b)
	}
	return r
}

func (r *RollingHash) Reset() {
	r.buffer = make([]byte, r.l)
	r.a = 0
	r.b = 0
	r.addOperationsCount = 0
}

func (r *RollingHash) Hash() uint32 {
	s := r.a + moduloVal*r.b
	return s
}

func (r RollingHash) Buffer() []byte {
	if len(r.buffer) > r.addOperationsCount {
		return append([]byte{}, r.buffer[len(r.buffer)-r.addOperationsCount:]...)
	}

	return append([]byte{}, r.buffer...)
}
