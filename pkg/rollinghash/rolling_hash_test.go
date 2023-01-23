package rollinghash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RollingHashHasCorrectValueForDifferentInputs(t *testing.T) {

	tests := []struct {
		name        string
		input       []byte
		expectedVal uint32
	}{
		{
			"returns correct value for zeros",
			[]byte{0, 0, 0, 0},
			0,
		},
		{
			"returns correct value for ones",
			[]byte{1, 1, 1, 1},
			655364,
		},
		{
			"returns correct value for ordered numbers",
			[]byte{1, 2, 3, 4},
			1310730,
		},
		{
			"returns correct value for max byte val",
			[]byte{255, 255, 255, 255},
			167117820,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := New(4).AddBuffer(test.input).Hash()
			assert.Equal(t, test.expectedVal, val)
		})
	}
}

func Test_AddInRollInChangesHash(t *testing.T) {
	input := []byte{48, 1, 15, 234}
	var toRollin byte = 186

	h := New(4).AddBuffer(input)
	h.Add(toRollin)

	assert.Equal(t, uint32(46072244), h.Hash())
}

func Test_RolledInHashShouldEqualInitialForSameData(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}
	var toRollin1 byte = 20
	input2 := append(input1[1:], toRollin1)

	h1 := New(4).AddBuffer(input1)

	h1.Add(toRollin1)
	h2 := New(4).AddBuffer(input2)

	assert.Equal(t, h2.Hash(), h1.Hash())

	var toRollin2 byte = 235
	input3 := append(input2[1:], toRollin2)

	h1.Add(toRollin2)
	h3 := New(4).AddBuffer(input3)

	assert.Equal(t, h3.Hash(), h1.Hash())
}

func Test_AddBufferHashShouldEqualInitialForSameData(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}
	var toRollin1 byte = 20
	var toRollin2 byte = 235

	h1 := New(4).AddBuffer(input1)
	h1.AddBuffer([]byte{toRollin1, toRollin2})

	input2 := append(input1[2:], toRollin1, toRollin2)

	h2 := New(4).AddBuffer(input2)

	assert.Equal(t, h2.Hash(), h1.Hash())
}

func Test_ReturnsDifferentValuesForDifferentInputs(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}
	input2 := []byte{10, 23, 34, 43}

	h1 := New(4).AddBuffer(input1)
	h2 := New(4).AddBuffer(input2)

	assert.NotEqual(t, h2.Hash(), h1.Hash())
}

func Test_ReturnsDifferentValuesAfterAddWasCalled(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}

	h1 := New(4).AddBuffer(input1)

	initHash := h1.Hash()
	h1.Add(23)

	assert.NotEqual(t, initHash, h1.Hash())
}

func Test_ReturnsDifferentValuesAfterAddBufferWasCalled(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}

	h1 := New(4).AddBuffer(input1)

	initHash := h1.Hash()
	h1.AddBuffer([]byte{23, 10, 15})

	secondHash := h1.Hash()
	assert.NotEqual(t, initHash, secondHash)
}
