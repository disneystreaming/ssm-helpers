package session

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// CheckSSMStartSession takes an SSM API session and instance ID and verifies whether or not the instance is available for start-session functionality
func CheckSSMStartSession(context ssmiface.SSMAPI, instanceID *string) (ready bool, err error) {
	// Create our start-session input object
	ssInput := &ssm.StartSessionInput{
		Target: instanceID,
	}

	// Call StartSession() to check the return value; if err is nil, start-session will work fine
	// If the error message is TargetNotConnected, it means that start-session will fail.
	output, err := context.StartSession(ssInput)

	// Call TerminateSession() to clear the start-session instance, as these are rate limited to 100 per account, per region
	if output.SessionId != nil {
		tsInput := &ssm.TerminateSessionInput{
			SessionId: output.SessionId,
		}

		_, err = context.TerminateSession(tsInput)
		if err != nil {
			return true, err
		}
	}

	if awsErr, ok := err.(awserr.Error); ok && err != nil && awsErr.Code() == "TargetNotConnected" {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
