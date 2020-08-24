package invocation

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	mocks "github.com/disneystreaming/ssm-helpers/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetResult(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockSSMClient{}

	successCmd, badCmd, mockInstance :=
		aws.String("success-id"), aws.String("bad-id"), aws.String("i-123")

	oc := make(chan *ssm.GetCommandInvocationOutput)
	ec := make(chan error)

	t.Run("valid ID", func(t *testing.T) {
		go GetResult(mockSvc, successCmd, mockInstance, oc, ec)
		select {
		case result := <-oc:
			assert.Equal("success-id", *result.CommandId)
		case err := <-ec:
			assert.Empty(err)
		}

	})

	t.Run("invalid ID", func(t *testing.T) {
		go GetResult(mockSvc, badCmd, mockInstance, oc, ec)
		select {
		case result := <-oc:
			assert.Empty(result)
		case err := <-ec:
			assert.Error(err)
		}
	})
}
