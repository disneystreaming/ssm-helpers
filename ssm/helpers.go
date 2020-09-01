package ssm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/disneystreaming/ssm-helpers/aws/session"
	ec2helpers "github.com/disneystreaming/ssm-helpers/ec2"
	"github.com/disneystreaming/ssm-helpers/ssm/instance"
	"github.com/disneystreaming/ssm-helpers/ssm/invocation"
	startsession "github.com/disneystreaming/ssm-helpers/ssm/session"
)

//CreateSSMDescribeInstanceInput returns an *ssm.DescribeInstanceInformationInput object for use when calling the DescribeInstanceInformation() method
func CreateSSMDescribeInstanceInput(filters map[string]string, instances CommaSlice) *ssm.DescribeInstanceInformationInput {
	var iisFilters []*ssm.InstanceInformationStringFilter

	//if you have multiple filter groups, this drops all but the last
	buildFilters(filters, &iisFilters)
	// Build our filters based on --filter and --instances flags

	if len(instances) > 0 {
		AppendSSMFilter(&iisFilters, NewSSMInstanceFilter("InstanceIds", instances))
	}

	ssmInput := &ssm.DescribeInstanceInformationInput{
		Filters: iisFilters,
	}

	// Max number of results per page allowed by the API is 50
	ssmInput.SetMaxResults(50)

	return ssmInput
}

func addInstanceInfo(instanceID *string, tags map[string]string, instancePool *instance.InstanceInfoSafe, profile string, region string) {
	instancePool.Lock()
	// If the instance is good, append its info to the master list
	instancePool.AllInstances[*instanceID] = instance.InstanceInfo{
		InstanceID: *instanceID,
		Profile:    profile,
		Region:     region,
		Tags:       tags,
	}
	instancePool.Unlock()
}

func checkInvocationStatus(client ssmiface.SSMAPI, commandID *string) (done bool, err error) {
	var invocation *ssm.ListCommandsOutput
	if invocation, err = client.ListCommands(&ssm.ListCommandsInput{
		CommandId: commandID,
	}); err != nil {
		return true, fmt.Errorf("Encountered an error when trying to call the ListCommands API with CommandId: %v\n%v", *commandID, err)
	}

	if len(invocation.Commands) != 1 {
		return true, fmt.Errorf("Incorrect number of invocations returned for given command ID; expected 1, got %d", len(invocation.Commands))
	}

	switch *invocation.Commands[0].Status {
	case "Pending", "InProgress":
		return false, nil
	default:
		return true, nil
	}
}

// RunInvocations invokes an SSM document with given parameters on the provided slice of instances
func RunInvocations(sess *session.Session, client ssmiface.SSMAPI, wg *sync.WaitGroup, input *ssm.SendCommandInput, results *invocation.ResultSafe) {
	defer wg.Done()
	var scOutput *ssm.SendCommandOutput
	var err error

	// Send our command input to SSM
	if scOutput, err = client.SendCommand(input); err != nil {
		sess.Logger.Error("Error when calling the SendCommand API")
		addError(results, sess, err)
		return
	}

	commandID := scOutput.Command.CommandId
	sess.Logger.Infof("Started invocation %v for %v in %v", *commandID, sess.ProfileName, *sess.Session.Config.Region)

	// Watch status of invocation to see when it's done and we can get the output
	for done := false; !done; time.Sleep(2 * time.Second) {
		if done, err = checkInvocationStatus(client, commandID); err != nil {
			addError(results, sess, err)
			return
		}
	}

	// Set up our channels and LCI input object
	oc := make(chan *ssm.GetCommandInvocationOutput)
	ec := make(chan error)
	lciInput := &ssm.ListCommandInvocationsInput{
		CommandId: commandID,
	}

	// Iterate through the details of the invocations returned
	if err = client.ListCommandInvocationsPages(
		lciInput,
		func(page *ssm.ListCommandInvocationsOutput, lastPage bool) bool {
			for _, entry := range page.CommandInvocations {
				// Fetch the results of our invocation for all provided instances
				go invocation.GetResult(client, commandID, entry.InstanceId, oc, ec)

				// Wait for results to return until the combined total of results and errors
				select {
				case result := <-oc:
					addInvocationResults(results, sess, result)
				case err := <-ec:
					addError(results, sess, err)
				}
			}

			// Last page, break out
			if page.NextToken == nil {
				return false
			}

			lciInput.SetNextToken(*page.NextToken)
			return true
		}); err != nil {
		sess.Logger.Error("Error when calling ListCommandInvocations API")
		addError(results, sess, err)
	}
}

func addInvocationResults(results *invocation.ResultSafe, session *session.Session, info ...*ssm.GetCommandInvocationOutput) {
	var newResults []*invocation.Result
	for _, v := range info {
		results.Add(&invocation.Result{
			InvocationResult: v,
			ProfileName:      session.ProfileName,
			Region:           *session.Session.Config.Region,
			Status:           invocation.Status(*v.StatusDetails),
		})
	}

	results.Lock()
	results.InvocationResults = append(results.InvocationResults, newResults...)
	results.Unlock()
}

func addError(results *invocation.ResultSafe, session *session.Session, err error) {
	results.Add(&invocation.Result{
		ProfileName: session.ProfileName,
		Region:      *session.Session.Config.Region,
		Status:      invocation.ClientError,
		Error:       err,
	})
}

// CheckInstanceReadiness iterates through a list of instances and verifies whether or not it is start-session capable. If it is, it appends the instance info to an instances.InstanceInfoSafe slice.
func CheckInstanceReadiness(session *session.Session, client ssmiface.SSMAPI, instanceList []*ssm.InstanceInformation, limit int, readyInstancePool *instance.InstanceInfoSafe) {
	var readyInstances, ec2Instances []*string
	var instanceCount int

	for _, instance := range instanceList {
		if instanceCount >= limit {
			continue
		}

		// Check and see if our instance supports start-session
		ready, err := startsession.CheckSessionReadiness(client, instance.InstanceId)
		if !ready && err != nil {
			session.Logger.Error(fmt.Errorf("Error when trying to check session readiness for instance %v\n%v", *instance.InstanceId, err))
			return
		}

		// Instances that are verified as being ready for sessions
		readyInstances = append(readyInstances, instance.InstanceId)

		// EC2 instances are all non-managed, so let's create a slice of instances that have fetchable tags
		if !strings.HasPrefix(*instance.InstanceId, "mi-") {
			ec2Instances = append(ec2Instances, instance.InstanceId)
		}

		instanceCount++
	}

	// If the instance is good, let's get the tags to display during instance selection
	ec2Client := ec2.New(session.Session)
	tags, err := ec2helpers.GetEC2InstanceTags(ec2Client, ec2Instances)
	if err != nil {
		session.Logger.Error(err)
	}

	if limit > len(readyInstances) {
		limit = len(readyInstances)
	}

	for _, i := range readyInstances[:limit] {
		// Append our instance info to the master list
		addInstanceInfo(i, tags[*i], readyInstancePool, session.ProfileName, *session.Session.Config.Region)
	}
}
