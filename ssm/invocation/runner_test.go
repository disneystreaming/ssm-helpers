package invocation

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"

	mocks "github.com/disneystreaming/ssm-helpers/testing"
)

func TestRunSSMCommand(t *testing.T) {
	assert := assert.New(t)
	resultChan := make(chan *ssm.SendCommandOutput)
	errChan := make(chan error)

	// Set up our mocks for testing RunSSMCommand()
	mockSvc := &mocks.MockSSMClient{}
	mockParams := &RunShellScriptParameters{
		"commands":         aws.StringSlice([]string{"echo foo", "uname -a"}),
		"executionTimeout": aws.StringSlice([]string{"600"}),
	}

	t.Run("dry run flag false", func(t *testing.T) {
		go RunSSMCommand(mockSvc, mockParams, false, resultChan, errChan, "i-123456", "i-654321")
		output, _ := <-resultChan, <-errChan

		// Do we have two instance IDs in our output?
		assert.Len(output.Command.InstanceIds, 2, "Incorrect nunber of instances passed to RunSSMCommand()")

		// Did we submit two separate commands to be executed by this document?
		assert.Len(output.Command.Parameters["commands"], 2, "Incorrect number of commands passed from mockParams object into SendCommandInput")
	})

	t.Run("dry run flag true", func(t *testing.T) {
		go RunSSMCommand(mockSvc, mockParams, true, resultChan, errChan, "i-123456", "i-654321")
		output, _ := <-resultChan, <-errChan

		assert.Nil(output, "Dry run flag enabled, should not have received any output")
	})
}

func TestGetCommandInvocationResult(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockSSMClient{}

	// Mock two batches with two instances each
	jobs := []*ssm.SendCommandOutput{
		{
			Command: &ssm.Command{
				CommandId:    aws.String("123456123456123456123456123456123456"),
				DocumentName: aws.String("AWS-RunShellScript"),
				InstanceIds:  aws.StringSlice([]string{"i-12345", "i-67890"}),
			},
		},
		{
			Command: &ssm.Command{
				CommandId:    aws.String("abcdefabcdefabcdefabcdefabcdefabcdef"),
				DocumentName: aws.String("AWS-RunShellScript"),
				InstanceIds:  aws.StringSlice([]string{"i-54321", "i-09876"}),
			},
		},
	}
	// So we can track state for each mock invocation in order to test handling of different return values
	for _, v := range jobs {
		for _, v := range v.Command.InstanceIds {
			os.Setenv(fmt.Sprintf("%s-trycount", *v), "0")
		}
	}
	results, err := GetCommandInvocationResult(mockSvc, jobs...)

	assert.Nilf(err, "Unexpected error returned from GetCommandInvocationResult\n%v", err)
	assert.Lenf(results, 4, "Incorrect number of invocation results returned; got %d, expected 4", len(results))
}
