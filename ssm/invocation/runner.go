package invocation

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	log "github.com/sirupsen/logrus"
)

// RunSSMCommand uses an SSM session, pre-defined SSM document parameters, the dry run flag, and any number of instance IDs and executes the given command
// using the AWS-RunShellScript SSM document. It returns an *ssm.SendCommandOutput object, which contains the execution ID of the command, which we use to
// check the progress/status of the invocation.
func RunSSMCommand(session ssmiface.SSMAPI, params *RunShellScriptParameters, dryRunFlag bool, resultChan chan *ssm.SendCommandOutput, errChan chan error, instanceID ...string) {
	var err error
	var output *ssm.SendCommandOutput

	ssmCommandInput := &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  aws.StringSlice(instanceID),
		Parameters:   *params}
	if !dryRunFlag {
		output, err = session.SendCommand(ssmCommandInput)
	}

	resultChan <- output
	errChan <- err
}

// GetCommandInvocationResult takes an SSM context and any number of *ssm.SendCommandOutput objects and iterates through them until the invocation is complete.
// Each invocation is checked concurrently, but the method as a whole is blocking until all invocations have returned a finishing result, whether successful or not.
func GetCommandInvocationResult(context ssmiface.SSMAPI, jobs ...*ssm.SendCommandOutput) (invocationStatus []*ssm.GetCommandInvocationOutput, err error) {
	// We're creating this here as well as in main() because otherwise we don't have the appropriate logging context
	errLog := log.New()
	errLog.SetFormatter(&log.TextFormatter{
		// Disable level truncation, timestamp, and pad out the level text to even it up
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	wg := sync.WaitGroup{}

	type resultsSafe struct {
		sync.Mutex
		results []*ssm.GetCommandInvocationOutput
	}

	var results resultsSafe

	// Concurrently iterate through all items in []instanceIDs and get the invocation status
	for _, v := range jobs {
		if v.Command != nil {
			for _, i := range v.Command.InstanceIds {
				wg.Add(1)
				go func(v *ssm.SendCommandOutput, i *string, context ssmiface.SSMAPI) {
					defer wg.Done()
					/*
						GetCommandInvocation() requires a GetCommandInvocationInput object, which
						has required parameters CommandId and InstanceId. It is important to note
						that unlike the execution of the command, you can only retrieve the invocation
						results for one instance+command at a time.
					*/
					gciInput := &ssm.GetCommandInvocationInput{
						CommandId:  v.Command.CommandId,
						InstanceId: i,
					}

					// Retrieve the status of the command invocation
					status, err := context.GetCommandInvocation(gciInput)

					// If we get "InvocationDoesNotExist", it just means we tried to check the results too quickly
					for awsErr, ok := err.(awserr.Error); ok && err != nil && awsErr.Code() == "InvocationDoesNotExist"; {
						time.Sleep(1000 * time.Millisecond)
						status, err = context.GetCommandInvocation(gciInput)
					}

					// If we somehow throw a real error here, something has gone screwy with our invocation or the target instance
					// See the docs on ssm.GetCommandInvocation() for error details
					if err != nil {
						errLog.Errorln(err)
						return
					}

					// If the invocation is in a pending state, we sleep for a couple seconds before retrying the query
					// NOTE: This may need to change based on API limits, but as there is no documentation, we'll have to wait and see.
					for *status.StatusDetails == "InProgress" || *status.StatusDetails == "Pending" {
						status, err = context.GetCommandInvocation(gciInput)
						time.Sleep(2000 * time.Millisecond)
					}

					if err != nil {
						errLog.Errorln(err)
						return
					}

					// Append the result to our slice of results
					results.Lock()
					results.results = append(results.results, status)
					results.Unlock()
				}(v, i, context)
			}
		}
	}

	wg.Wait()
	// Return
	return results.results, err
}
