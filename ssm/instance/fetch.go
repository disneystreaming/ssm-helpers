package instance

import (
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// getSSMInstances takes an SSM session and an *ssm.DescribeInstanceInformationInput object to get information about a given SSM-managed EC2 instance.
func getSSMInstances(session ssmiface.SSMAPI, input *ssm.DescribeInstanceInformationInput) (output *ssm.DescribeInstanceInformationOutput, err error) {
	return session.DescribeInstanceInformation(input)
}

// GetAllSSMInstances takes an SSM session and an *ssm.DescribeInstanceInformationInput object, and calls getSSMInstances() to query the SSM API for
// information about a given SSM-managed EC2 instance. It also filters out any instances returned that are unresponsive to ping or not running Linux.
// Filtering can only be done based on either instance information OR instance tags for any given instance of a query. As such, filtering of instances
// by things like bamazon and env tags must be done after retrieving the instance information.
func GetAllSSMInstances(session ssmiface.SSMAPI, input *ssm.DescribeInstanceInformationInput, checkLatestAgent bool) (output []*ssm.InstanceInformation, err error) {
	// First call to set up NextToken
	results, err := getSSMInstances(session, input)

	var allResults []*ssm.InstanceInformation
	for _, v := range results.InstanceInformationList {
		// Remove any instances that are unresponsive to ping or not running Linux
		if checkLatestAgent {
			if *v.PingStatus == "Online" && *v.PlatformType == "Linux" && *v.IsLatestVersion {
				allResults = append(allResults, v)
			}
		} else {
			if *v.PingStatus == "Online" && *v.PlatformType == "Linux" {
				allResults = append(allResults, v)
			}
		}
	}
	// Repeat call until we've paginated through all results
	for results.NextToken != nil {
		if *results.NextToken != "" {
			input.SetNextToken(*results.NextToken)
		}
		results, _ = getSSMInstances(session, input)
		for _, v := range results.InstanceInformationList {
			// Remove any instances are unresponsive to ping or not running Linux
			if checkLatestAgent {
				if *v.PingStatus == "Online" && *v.PlatformType == "Linux" && *v.IsLatestVersion {
					allResults = append(allResults, v)
				}
			} else {
				if *v.PingStatus == "Online" && *v.PlatformType == "Linux" {
					allResults = append(allResults, v)
				}
			}
		}
	}
	return allResults, err
}
