package ssm

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	log "github.com/sirupsen/logrus"

	"github.com/disneystreaming/ssm-helpers/aws/session"
	ec2helpers "github.com/disneystreaming/ssm-helpers/ec2"
	"github.com/disneystreaming/ssm-helpers/ssm/instance"
	"github.com/disneystreaming/ssm-helpers/ssm/invocation"
	startsession "github.com/disneystreaming/ssm-helpers/ssm/session"
)

//CreateSSMDescribeInstanceInput returns an *ssm.DescribeInstanceInformationInput object for use when calling the DescribeInstanceInformation() method
func CreateSSMDescribeInstanceInput(filters []map[string]string, instances CommaSlice) *ssm.DescribeInstanceInformationInput {
	var ssmInputFilters []*ssm.InstanceInformationStringFilter

	if len(filters) > 0 {
		for _, v := range filters {
			//if you have multiple filter groups, this drops all but the last
			BuildFilters(v, &ssmInputFilters)
			// Build our filters based on --filter and --instances flags
		}
	}

	if len(instances) > 0 {
		AppendSSMFilter(&ssmInputFilters, NewSSMInstanceFilter("InstanceIds", instances))
	}

	ssmInput := &ssm.DescribeInstanceInformationInput{
		Filters: ssmInputFilters,
	}

	// Max number of results per page allowed by the API is 50
	ssmInput.SetMaxResults(50)

	return ssmInput
}

func addInstanceInfo(instanceID *string, tags []ec2helpers.InstanceTags, instancePool *instance.InstanceInfoSafe, profile string, region string) {
	for _, v := range tags {
		instancePool.Lock()
		// If the instance is good, append its info to the master list
		instancePool.AllInstances[*instanceID] = instance.InstanceInfo{
			InstanceID: *instanceID,
			Profile:    profile,
			Region:     region,
			Tags:       v.Tags,
		}
		instancePool.Unlock()
	}
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
func RunInvocations(sess *session.Pool, client ssmiface.SSMAPI, wg *sync.WaitGroup, input *ssm.SendCommandInput, results *invocation.ResultSafe) {
	defer wg.Done()

	oc := make(chan *ssm.GetCommandInvocationOutput)
	ec := make(chan error)
	var scOutput *ssm.SendCommandOutput
	var err error

	// Send our command input to SSM
	if scOutput, err = client.SendCommand(input); err != nil {
		sess.Logger.Errorf("Error when calling the SendCommand API for account %v in %v\n%v", sess.ProfileName, *sess.Session.Config.Region, err)
		return
	}

	commandID := scOutput.Command.CommandId
	sess.Logger.Infof("Started invocation %v for %v in %v", *commandID, sess.ProfileName, *sess.Session.Config.Region)

	// Watch status of invocation to see when it's done and we can get the output
	for done := false; !done; time.Sleep(2 * time.Second) {
		if done, err = checkInvocationStatus(client, commandID); err != nil {
			sess.Logger.Error(err)
			return
		}
	}

	// Set up our LCI input object
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
					sess.Logger.Error(err)
				}
			}

			// Last page, break out
			if page.NextToken == nil {
				return false
			}

			lciInput.SetNextToken(*page.NextToken)
			return true
		}); err != nil {
		sess.Logger.Error(fmt.Errorf("Error when calling ListCommandInvocations API\n%v", err))
	}

}

func addInvocationResults(results *invocation.ResultSafe, session *session.Pool, info ...*ssm.GetCommandInvocationOutput) {
	for _, v := range info {
		var result = &invocation.Result{
			InvocationResult: v,
			ProfileName:      session.ProfileName,
			Region:           *session.Session.Config.Region,
			Status:           *v.StatusDetails,
		}

		results.Lock()
		results.InvocationResults = append(results.InvocationResults, result)
		results.Unlock()
	}
}

// CheckInstanceReadiness iterates through a list of instances and verifies whether or not it is start-session capable. If it is, it appends the instance info to an instances.InstanceInfoSafe slice.
func CheckInstanceReadiness(sp *session.Pool, ssmSession *ssm.SSM, instanceList []*ssm.InstanceInformation, readyInstancePool *instance.InstanceInfoSafe, limit int) {
	var readyInstances int
	ec2Sess := ec2.New(sp.Session)

	for _, instance := range instanceList {
		if readyInstances < limit {
			// Check and see if our instance supports start-session
			ready, err := startsession.CheckSSMStartSession(ssmSession, instance.InstanceId)
			if !ready {
				if err != nil {
					log.Debug(err)
				}
				continue
			}

			// If the instance is good, let's get the tags to display during instance selection
			tags, err := ec2helpers.GetEC2InstanceTags(ec2Sess, *instance.InstanceId)
			if err != nil {
				log.Errorf("Could not retrieve tags for instance %s\n%s", *instance.InstanceId, err)
				continue
			}

			// Append our instance info to the master list
			addInstanceInfo(instance.InstanceId, tags, readyInstancePool, sp.ProfileName, *sp.Session.Config.Region)
		}
		readyInstances++
	}
}

// GetInstanceList creates a DescribeInstanceInput object and returns all SSM instances that match the provided filters
func GetInstanceList(ssmSession *ssm.SSM, filters []map[string]string, instanceInput CommaSlice, checkLatestAgent bool, infoChan chan []*ssm.InstanceInformation, errChan chan error) {

	//Create our instance input object (filters, instances)
	ssmInput := CreateSSMDescribeInstanceInput(filters, instanceInput)

	// Get our list of all instances in the current session that match
	// the filters we've configured (this will also pare down instances)
	output, err := instance.GetAllSSMInstances(ssmSession, ssmInput, checkLatestAgent)

	infoChan <- output
	/*
		This can happen when a given profile permutation doesn't have the
		correct permissions or when it lacks SSM access. Since we may have
		multiple sessions to iterate through and the rest of the program
		does nothing without an instance as input, this is a non-fatal error.
	*/
	errChan <- err

}
