package ssm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	assert := assert.New(t)

	t.Run("listslice set()", func(t *testing.T) {
		ls := ListSlice{"foo"}
		ls.Set("bar baz")

		assert.Lenf(ls, 2, "Set returned wrong sized slice, got %d, expected 2.", len(ls))
	})

	t.Run("semislice set()", func(t *testing.T) {
		ss := SemiSlice{"foo"}
		ss.Set("bar;baz")

		assert.Lenf(ss, 3, "Set returned wrong sized slice, got %d, expected 3", len(ss))
	})

	t.Run("commaslice set()", func(t *testing.T) {
		cs := CommaSlice{"foo"}
		cs.Set("bar,baz")

		assert.Lenf(cs, 3, "Set returned wrong sized slice, got %d, expected 3", len(cs))
	})
}

func TestString(t *testing.T) {
	assert := assert.New(t)

	t.Run("listslice string()", func(t *testing.T) {
		ls := ListSlice{"foo", "bar"}

		assert.Equalf(ls.String(), "[foo bar]", "Method returned incorrect string representation, got %s, expected [foo bar]", ls.String())
	})

	t.Run("semislice string()", func(t *testing.T) {
		ss := SemiSlice{"foo", "bar"}

		assert.Equalf(ss.String(), "[foo bar]", "Method returned incorrect string representation, got %s, expected [foo bar]", ss.String())
	})

	t.Run("commaslice string()", func(t *testing.T) {
		cs := CommaSlice{"foo", "bar"}

		assert.Equalf(cs.String(), "[foo bar]", "Method returned incorrect string representation, got %s, expected [foo bar]", cs.String())
	})
}
