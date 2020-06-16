package ssm

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/go-multierror"
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

// RunInvocations invokes an SSM document with given parameters on the provided slice of instances
func RunInvocations(sp *session.Pool, sess *ssm.SSM, instances []*ssm.InstanceInformation, params *invocation.RunShellScriptParameters, dryRun bool, resultsPool *invocation.ResultSafe) (err error) {
	var commandOutput invocation.CommandOutputSafe
	var invError error

	scoChan := make(chan *ssm.SendCommandOutput)
	errChan := make(chan error)

	/*
		In a standard deployment, SSM allows us run commands on a maximum of
		up to 50 instances simultaneously.

		(Technically, it does an exponential deployment, where it deploys to n^2
		instances at a time (up to 50), where n is the last number of instances
		on which the command completed.)

		To speed up execution, we can split the instances into arbitrarily-sized
		batches and run the command on every batch concurrently. In the current
		implementation, we are effectively using a batch size of 1 for maximum
		concurrency.
	*/

	for _, instance := range instances {
		go invocation.RunSSMCommand(sess, params, dryRun, scoChan, errChan, *instance.InstanceId)
		output, err := <-scoChan, <-errChan

		if err != nil {
			invError = multierror.Append(invError, err)
		}

		if output != nil {
			addInvocationInfo(output, &commandOutput)
		}
	}

	// Fetch the results of our invocation for all provided instances
	invocationStatus, err := invocation.GetCommandInvocationResult(sess, commandOutput.Output...)
	if err != nil {
		// If we somehow throw an error here, something has gone screwy with our invocation or the target instance
		// See the docs on ssm.GetCommandInvocation() for error details
		invError = multierror.Append(invError, err)
	}

	// Iterate through all retrieved invocation results to add some extra context
	addInvocationResults(invocationStatus, resultsPool, sp)
	return invError
}

func addInvocationInfo(info *ssm.SendCommandOutput, infoPool *invocation.CommandOutputSafe) {
	if info != nil {
		infoPool.Lock()
		infoPool.Output = append(infoPool.Output, info)
		infoPool.Unlock()
	}

}

func addInvocationResults(info []*ssm.GetCommandInvocationOutput, results *invocation.ResultSafe, session *session.Pool) {
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
