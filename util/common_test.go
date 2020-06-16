package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommaSplit(t *testing.T) {
	assert := assert.New(t)

	args := CommaSplit("foo,bar,baz")
	assert.Lenf(args, 3, "Incorrect number of strings returned, got %d, expected 3", len(args))
}

func TestSemiSplit(t *testing.T) {
	assert := assert.New(t)

	args := SemiSplit("foo;bar;baz")
	assert.Lenf(args, 3, "Incorrect number of strings returned, got %d, expected 3", len(args))
}

func TestSliceToMap(t *testing.T) {
	assert := assert.New(t)

	// Initialize test data
	testSlice := []string{
		"foo=1", "bar=2", "baz=3",
	}

	// Create test map
	testMap := make(map[string]string)

	// This should result in testMap containing 3 k=v mappings
	SliceToMap(testSlice, &testMap)
	assert.Lenf(testMap, 3, "Finished map contains the incorrect number of entries; got %d, expected 3", len(testMap))

	testSlice = []string{
		"ping=4", "pong=5",
	}

	// Test appending entries to existing map
	SliceToMap(testSlice, &testMap)
	assert.Lenf(testMap, 5, "Appended map contains the incorrect number of entries; got %d, expected 5", len(testMap))

	testSlice = []string{
		"foo=6", "ping=7",
	}

	// Test overwriting of entries in existing map
	SliceToMap(testSlice, &testMap)

	assert.Lenf(testMap, 5, "Appended map contains the incorrect number of entries; got %d, expected 5", len(testMap))
	assert.Truef(
		testMap["foo"] == "6" && testMap["ping"] == "7",
		"Entries that should have been overwritten contain incorrect values: returned foo=%s, ping=%s, expected foo=6, ping=7",
		testMap["foo"],
		testMap["ping"],
	)

}
