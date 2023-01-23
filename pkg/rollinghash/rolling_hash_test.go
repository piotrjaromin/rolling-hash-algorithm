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
			val := InitRollingHash(test.input).Hash()
			assert.Equal(t, test.expectedVal, val)
		})
	}
}

func Test_AddInRollInChangesHash(t *testing.T) {
	input := []byte{48, 1, 15, 234}
	var toRollin byte = 186

	h := InitRollingHash(input)
	h.Add(toRollin)

	assert.Equal(t, uint32(46072244), h.Hash())
}

func Test_RolledInHashShouldEqualInitialForSameData(t *testing.T) {
	input1 := []byte{34, 23, 34, 234}
	var toRollin1 byte = 20
	input2 := append(input1[1:], toRollin1)

	h1 := InitRollingHash(input1)

	h1.Add(toRollin1)
	h2 := InitRollingHash(input2)

	assert.Equal(t, h2.Hash(), h1.Hash())

	var toRollin2 byte = 235
	input3 := append(input2[1:], toRollin2)

	h1.Add(toRollin2)
	h3 := InitRollingHash(input3)

	assert.Equal(t, h3.Hash(), h1.Hash())
}
