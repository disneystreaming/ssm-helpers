package mocks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type MockSSMClient struct {
	ssmiface.SSMAPI
}

func (m *MockSSMClient) StartSession(input *ssm.StartSessionInput) (output *ssm.StartSessionOutput, err error) {

	// Simulate a working SSM instance that supports start-session
	if *input.Target == "i-123" {
		return &ssm.StartSessionOutput{
			SessionId: aws.String("ready-instance-id"),
		}, nil
	}

	// Simulate an SSM instance with bad permissions
	if *input.Target == "i-456" {
		return &ssm.StartSessionOutput{}, awserr.New("TargetNotConnected", "bad instance role permissions make this fail on ssm-managed instances", nil)
	}

	// Simulate a method call that fails for an arbitrary, non-TargetNotConnected reason
	if *input.Target == "i-789" {
		return &ssm.StartSessionOutput{}, fmt.Errorf("This represents any error other than TargetNotConnected.")
	}

	// Simulate a working SSM instance that supports start-session but then fails when TerminateSession() is called
	if *input.Target == "i-000" {
		return &ssm.StartSessionOutput{
			SessionId: aws.String("session-term-error"),
		}, nil
	}

	return
}

func (m *MockSSMClient) TerminateSession(input *ssm.TerminateSessionInput) (output *ssm.TerminateSessionOutput, err error) {
	if *input.SessionId == "session-term-error" {
		return &ssm.TerminateSessionOutput{
			SessionId: input.SessionId,
		}, awserr.New("DoesNotExistException", "this tends to occur when you hit rate limits", nil)
	}

	return &ssm.TerminateSessionOutput{
		SessionId: input.SessionId,
	}, nil
}

func (m *MockSSMClient) GetCommandInvocation(input *ssm.GetCommandInvocationInput) (output *ssm.GetCommandInvocationOutput, err error) {
	var status string

	switch *input.CommandId {
	case "success-id":
		status = "Success"
	case "failed-id", "mixed-id":
		status = "Failed"
	case "pending-id":
		status = "Pending"
	case "bad-id":
		return nil, awserr.New("InvalidCommandId", "InvalidCommandId", nil)
		}

		output = &ssm.GetCommandInvocationOutput{
			InstanceId:    input.InstanceId,
			CommandId:     input.CommandId,
		StatusDetails: aws.String(status),
		}

		return output, nil
	}

func (m *MockSSMClient) SendCommand(input *ssm.SendCommandInput) (output *ssm.SendCommandOutput, err error) {
	if len(input.InstanceIds) > 0 && len(input.Targets) > 0 {
		return nil, fmt.Errorf("Cannot specify instance IDs and SSM targets in same SendCommandInput")
	}

	// Mock our response from the SSM API
	output = &ssm.SendCommandOutput{
		Command: &ssm.Command{
			CommandId:    aws.String("1234561234561234561234561235456"),
			DocumentName: input.DocumentName,
			InstanceIds:  input.InstanceIds,
			Parameters:   input.Parameters,
			Targets:      input.Targets,
		},
	}
	return output, nil
}

func (m *MockSSMClient) DescribeInstanceInformation(input *ssm.DescribeInstanceInformationInput) (output *ssm.DescribeInstanceInformationOutput, err error) {

	// Mock our response from the SSM API

	if input.NextToken == nil {
		output = &ssm.DescribeInstanceInformationOutput{
			InstanceInformationList: []*ssm.InstanceInformation{
				{
					PlatformType:    aws.String("Linux"),
					PingStatus:      aws.String("Offline"),
					InstanceId:      aws.String("i-23456"),
					IsLatestVersion: aws.Bool(true),
				},
				{
					PlatformType:    aws.String("Linux"),
					PingStatus:      aws.String("Online"),
					InstanceId:      aws.String("i-45678"),
					IsLatestVersion: aws.Bool(true),
				},
				{
					PlatformType:    aws.String("Windows"),
					PingStatus:      aws.String("Offline"),
					InstanceId:      aws.String("i-78901"),
					IsLatestVersion: aws.Bool(true),
				},
				{
					PlatformType:    aws.String("Linux"),
					PingStatus:      aws.String("Online"),
					InstanceId:      aws.String("i-98765"),
					IsLatestVersion: aws.Bool(false),
				},
			},
			NextToken: aws.String("eyJNYXJrZXIiOiBudWxsLCAiYm90b190cnVuY2F0ZV9hbW91bnQiOiAxfQ=="),
		}
		return output, err
	}

	output = &ssm.DescribeInstanceInformationOutput{
		InstanceInformationList: []*ssm.InstanceInformation{
			{
				PlatformType:    aws.String("Linux"),
				PingStatus:      aws.String("Online"),
				InstanceId:      aws.String("i-12345"),
				IsLatestVersion: aws.Bool(true),
			},
			{
				PlatformType:    aws.String("Linux"),
				PingStatus:      aws.String("Online"),
				InstanceId:      aws.String("i-34567"),
				IsLatestVersion: aws.Bool(true),
			},
			{
				PlatformType:    aws.String("Windows"),
				PingStatus:      aws.String("Online"),
				InstanceId:      aws.String("i-67890"),
				IsLatestVersion: aws.Bool(true),
			},
		},
		NextToken: nil,
	}

	return output, err

}
