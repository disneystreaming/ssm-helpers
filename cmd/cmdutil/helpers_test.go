package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateIntOrPercantageValue(t *testing.T) {
	assert := assert.New(t)

	t.Run("No input is false", func(t *testing.T) {
		expected := false
		actual := ValidateIntOrPercantageValue("")
		assert.Equal(expected, actual)
	})

	t.Run("Input len greater than 7 char is false", func(t *testing.T) {
		expected := false
		actual1 := ValidateIntOrPercantageValue("12345678")
		assert.Equal(expected, actual1)

		actual2 := ValidateIntOrPercantageValue("1234567%")
		assert.Equal(expected, actual2)
	})

	t.Run("Incorrect format is false", func(t *testing.T) {
		expected := false
		actual := ValidateIntOrPercantageValue("10%10")
		assert.Equal(expected, actual)
	})

	t.Run("Number is true", func(t *testing.T) {
		expected := true
		actual := ValidateIntOrPercantageValue("10")
		assert.Equal(expected, actual)
	})

	t.Run("Percentage is true", func(t *testing.T) {
		expected := true
		actual := ValidateIntOrPercantageValue("10%")
		assert.Equal(expected, actual)
	})
}
