package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	items := []int{12, 22, 53, 24, 75, 96, 67, 18, 39, 10, 13, 99, 31}

	counter := 0
	fn := func(min int, max int) (bool, error) {
		assert.Equal(t, counter, min)

		if max == len(items) {
			assert.Equal(t, len(items), max)
		} else {
			assert.Equal(t, counter+2, max)
		}

		counter += 2
		return true, nil
	}

	err := Chunk(len(items), 2, fn)
	assert.NoError(t, err)
}
