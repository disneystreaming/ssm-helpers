package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAWSProfiles(t *testing.T) {
	assert := assert.New(t)

	t.Run("test fake file", func(t *testing.T) {
		// Make sure that we fail gracefully when unable to access a given file
		_, err := GetAWSProfiles("fake")
		assert.NotNil(err, "Function should return an error when attempting to access a file that does not exist.")
	})

	t.Run("check for multiple paths", func(t *testing.T) {
		_, err := GetAWSProfiles("foo", "bar", "baz")
		assert.NotNil(err, "Function did not return an error when provided multiple paths.")
	})

	t.Run("test valid file", func(t *testing.T) {
		_, err := GetAWSProfiles("../testing/mock_aws_profiles")

		// Make sure we can access our test data
		assert.Nilf(err, "Error when trying to open test AWS profile data in ../testing/mock_aws_profiles\n%v", err)
	})

	t.Run("verify file contents", func(t *testing.T) {
		profiles, _ := GetAWSProfiles("../testing/mock_aws_profiles")

		// Verify that the correct number of results are being parsed out of our test data.
		assert.Lenf(profiles, 3, "Incorrect number of results parsed from test data, got %d, expected 3", len(profiles))

		// Validate that all of the entries exist in our results
		expectedElements := []string{"test-account-1", "test-account-2", "test-account-3"}
		assert.ElementsMatchf(profiles, expectedElements, "Incorrect values from parsed data, got %s, expected 'test-account-1', 'test-account-2', and 'test-account-3'.", profiles)
	})
}
