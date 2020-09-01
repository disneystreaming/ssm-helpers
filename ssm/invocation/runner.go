package invocation

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func GetTargets(client ssmiface.SSMAPI, commandID *string) (targets []*string, err error) {
	var out *ssm.ListCommandInvocationsOutput

	// Try a few times to get the invocation data, because it takes a little bit to have any information
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		if out, err = client.ListCommandInvocations(&ssm.ListCommandInvocationsInput{
			CommandId: commandID,
		}); err != nil {
			return nil, err
		}

		if len(out.CommandInvocations) > 0 {
			break
		}
	}

	if len(out.CommandInvocations) == 0 {
		return nil, fmt.Errorf("API response contained no invocations")
	}

	for _, inv := range out.CommandInvocations {
		targets = append(targets, inv.InstanceId)
	}

	return targets, nil
}

func GetResult(client ssmiface.SSMAPI, commandID *string, instanceID *string, gci chan *ssm.GetCommandInvocationOutput, ec chan error) {
	status, err := client.GetCommandInvocation(&ssm.GetCommandInvocationInput{
		CommandId:  commandID,
		InstanceId: instanceID,
	})

	switch {
	case err != nil:
		ec <- fmt.Errorf(
			`Error when calling GetCommandInvocation API with args:\n
			CommandId: %v\n
			InstanceId: %v\n%v`,
			*commandID, *instanceID, err)
	case status != nil:
		gci <- status
	}
}
