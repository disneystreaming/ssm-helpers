package session

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	mocks "github.com/disneystreaming/ssm-helpers/testing"
)

func TestCheckSSMStartSession(t *testing.T) {
	assert := assert.New(t)
	sess := &mocks.MockSSMClient{}

	t.Run("instance ready for start-session", func(t *testing.T) {
		ready, err := CheckSSMStartSession(sess, aws.String("i-123"))

		assert.Nil(err, "Method returned an error when given a working instance.")
		assert.Truef(ready, "Method returned %v, expected true", ready)
	})

	t.Run("instance with bad permissions", func(t *testing.T) {
		ready, err := CheckSSMStartSession(sess, aws.String("i-456"))

		assert.Nil(err, "Method did not handle 'TargetNotConnected' error properly.")
		assert.Falsef(ready, "Method returned %v, expected false", ready)
	})

	t.Run("non-permission related issue", func(t *testing.T) {
		ready, err := CheckSSMStartSession(sess, aws.String("i-789"))

		assert.NotNil(err, "Method did not return an error when a non-TargetNotConnected error occured.")
		assert.Falsef(ready, "Method returned %v, expected false", ready)
	})

	t.Run("instance fails terminate-session call", func(t *testing.T) {
		ready, err := CheckSSMStartSession(sess, aws.String("i-000"))

		assert.NotNil(err, "Method did not return an error when session termination failed.")
		assert.Truef(ready, "Method returned %v when session termination failed.", ready)
	})
}
