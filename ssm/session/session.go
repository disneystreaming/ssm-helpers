package session

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// CheckSessionReadiness takes an SSM API session and instance ID and verifies whether or not the instance is available for start-session functionality
func CheckSessionReadiness(context ssmiface.SSMAPI, instanceID *string) (connected bool, err error) {
	// Create our getConnectionStatus input object
	gcsInput := &ssm.GetConnectionStatusInput{
		Target: instanceID,
	}

	// Call GetConnectionStatus to determine if the given instance is ready for a session
	output, err := context.GetConnectionStatus(gcsInput)

	// Suppress errors due to rate limiting
	if awsErr, ok := err.(awserr.Error); ok && err != nil && awsErr.Code() == "ThrottlingException" {
		return false, nil
	}

	if err != nil || *output.Status == "NotConnected" {
		return false, err
	}

	return true, nil
}
