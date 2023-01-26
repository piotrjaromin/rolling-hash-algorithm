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

func Test_IfCStartedWithDifferentValuesConvergesForSameInput(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}
	input2 := []byte{111, 3, 55, 35}

	commonBytes := []byte{1, 2, 3, 4}

	h1 := New(4).AddBuffer(input1)
	h2 := New(4).AddBuffer(input2)

	for _, b := range commonBytes {
		h1.Add(b)
		h2.Add(b)
	}

	assert.Equal(t, h1.Hash(), h2.Hash())
}

func Test_BufferHasCorrectOrder(t *testing.T) {
	input1 := []byte{34, 23, 82, 234}

	h1 := New(4).AddBuffer(input1)
	assert.Equal(t, input1, h1.Buffer())

	var input2 byte = 54
	h1.Add(input2)

	expected2 := append(input1[1:], input2)
	assert.Equal(t, expected2, h1.Buffer())

	var input3 byte = 89
	h1.Add(input3)

	expected3 := append(expected2[1:], input3)
	assert.Equal(t, expected3, h1.Buffer())

	input4 := []byte{11, 76, 221}
	h1.AddBuffer(input4)

	expected4 := append(expected3[3:], input4...)
	assert.Equal(t, expected4, h1.Buffer())

	input5 := []byte{114, 169}
	h1.AddBuffer(input5)

	expected5 := append(expected4[2:], input5...)
	assert.Equal(t, expected5, h1.Buffer())
}
