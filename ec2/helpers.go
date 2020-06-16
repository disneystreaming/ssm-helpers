package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func getEC2InstanceInfo(context ec2iface.EC2API, instances ...string) (output []*ec2.DescribeInstancesOutput, err error) {

	instanceInput := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instances),
	}

	results, err := context.DescribeInstances(instanceInput)
	if err != nil {
		return nil, err
	}
	output = append(output, results)

	for results.NextToken != nil {
		if *results.NextToken != "" {
			instanceInput.SetNextToken(*results.NextToken)
		}
		// Get our next page of results
		results, err = context.DescribeInstances(instanceInput)
		if err != nil {
			return output, err
		}
		output = append(output, results)
	}

	return output, nil
}

// GetEC2InstanceTags accepts any number of instance strings and returns a populated InstanceTags{} object for each instance
func GetEC2InstanceTags(context ec2iface.EC2API, instances ...string) (tags []InstanceTags, err error) {

	instanceInfo, err := getEC2InstanceInfo(context, instances...)
	if err != nil {
		return nil, err
	}

	for _, page := range instanceInfo {
		for _, res := range page.Reservations {
			for _, inst := range res.Instances {

				tagStruct := &InstanceTags{
					Tags:       make(map[string]string),
					InstanceID: *inst.InstanceId,
				}

				for _, tag := range inst.Tags {
					tagStruct.Tags[*tag.Key] = *tag.Value
				}

				tags = append(tags, *tagStruct)
			}
		}
	}

	return tags, nil
}
