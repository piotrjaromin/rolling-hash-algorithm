package sync

const moduloVal uint32 = 1 << 16

type rollingHash struct {
	buffer []byte
	a      uint32
	b      uint32
	l      uint32
}

func InitRollingHash(data []byte) *rollingHash {
	l := len(data)

	var a uint32 = 0
	var b uint32 = 0

	for i, val := range data {
		a += uint32(val)
		b += uint32(l-i+1) * uint32(val)
	}

	a = a % moduloVal
	b = b % moduloVal

	buffer := make([]byte, 0, len(data))
	copy(buffer, data)

	return &rollingHash{
		buffer: buffer,
		a:      uint32(a),
		b:      uint32(b),
		l:      uint32(l),
	}
}

func (r *rollingHash) Add(b byte) {

	r.a = (r.a - uint32(r.buffer[0]) + uint32(b)) % moduloVal
	r.b = (r.b - r.l*uint32(r.buffer[0]) + r.a) % moduloVal

	for i := 0; i < int(r.l)-1; i++ {
		r.buffer[i] = r.buffer[i+1]
	}

	r.buffer[r.l-1] = b
}

func (r rollingHash) Hash() uint32 {
	s := r.a + moduloVal*r.b
	return s
}
