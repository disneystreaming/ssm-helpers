package instance

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

func GetSessionInstances(client ssmiface.SSMAPI, diiInput *ssm.DescribeInstanceInformationInput) (output []*ssm.InstanceInformation, err error) {
	// Fetch all instances that match the provided filters
	if err = client.DescribeInstanceInformationPages(
		diiInput,
		func(page *ssm.DescribeInstanceInformationOutput, lastPage bool) bool {
			for _, instance := range page.InstanceInformationList {
				output = append(output, instance)
			}

			// Last page, break out
			if page.NextToken == nil {
				return false
			}

			// If not, set the token in order to fetch the next page
			diiInput.SetNextToken(*page.NextToken)
			return true
		}); err != nil {
		return nil, fmt.Errorf("Could not retrieve SSM instance info\n%v", err)
	}

	return output, err
}
