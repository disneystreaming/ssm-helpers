package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMaxConcurrency(t *testing.T) {
	assert := assert.New(t)

	t.Run("No input is false", func(t *testing.T) {
		actual := ValidateMaxConcurrency("")
		assert.False(actual)
	})

	t.Run("Input len greater than 7 char is false", func(t *testing.T) {
		actual1 := ValidateMaxConcurrency("12345678")
		assert.False(actual1)

		actual2 := ValidateMaxConcurrency("1234567%")
		assert.False(actual2)
	})

	t.Run("Incorrect format is false", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10%10")
		assert.False(actual)
	})

	t.Run("Number is true", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10")
		assert.True(actual)
	})

	t.Run("Percentage is true", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10%")
		assert.True(actual)
	})
}

func TestValidateMaxErrors(t *testing.T) {
	assert := assert.New(t)

	t.Run("No input is false", func(t *testing.T) {
		actual := ValidateMaxErrors("")
		assert.False(actual)
	})

	t.Run("Input len greater than 7 char is false", func(t *testing.T) {
		actual1 := ValidateMaxErrors("12345678")
		assert.False(actual1)

		actual2 := ValidateMaxErrors("1234567%")
		assert.False(actual2)
	})

	t.Run("Incorrect format is false", func(t *testing.T) {
		actual := ValidateMaxErrors("10%10")
		assert.False(actual)
	})

	t.Run("Number is true", func(t *testing.T) {
		actual := ValidateMaxErrors("10")
		assert.True(actual)
	})

	t.Run("Percentage is true", func(t *testing.T) {
		actual := ValidateMaxErrors("10%")
		assert.True(actual)
	})
}
