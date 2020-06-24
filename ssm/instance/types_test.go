package instance

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	assert := assert.New(t)

	is := InstanceSlice{}
	is.Set("i-123,i-234,i-345")

	assert.Lenf(is, 3, "Set method returned an incorrect number of results, got %d, expected 3", len(is))
}

func TestString(t *testing.T) {
	assert := assert.New(t)

	is := InstanceSlice{"foo", "bar", "baz"}

	assert.Equalf(is.String(), "[foo bar baz]", "String() returned an improper representation of the InstanceSlice object, got %v, expected '[foo bar baz]'", is.String())
}

func TestFormatString(t *testing.T) {
	assert := assert.New(t)

	ii := InstanceInfo{
		InstanceID: "i-123",
		Region:     "us-east-1",
		Profile:    "test",
		Tags: map[string]string{
			"foo":     "oof",
			"bar":     "longvalue",
			"longkey": "baz",
		},
	}

	// Set up a regex match to check the number of
	rxPattern := regexp.MustCompile(`\S+\s+\S+\s+\S+\s+\S+\s+\S+`)

	assert.Truef(
		rxPattern.MatchString(ii.FormatString("foo", "bar")),
		"Incorrect string formatting in output.\nGot:\t\t'%v'\nExpected:\t'i-123  us-east-1       test    oof     baz'",
		ii.FormatString("foo", "longkey"),
	)
}
