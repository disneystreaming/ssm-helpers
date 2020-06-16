package mocks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type MockEC2Client struct {
	ec2iface.EC2API
}

func (m *MockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (output *ec2.DescribeInstancesOutput, err error) {

	// Set up our sample data for a few instances
	output = &ec2.DescribeInstancesOutput{
		NextToken: nil,
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						InstanceId: aws.String("i-123"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("id_foo"),
								Value: aws.String("321"),
							},
							{
								Key:   aws.String("env"),
								Value: aws.String("dev"),
							},
						},
					},
					{
						InstanceId: aws.String("i-456"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("id_bar"),
								Value: aws.String("654"),
							},
							{
								Key:   aws.String("env"),
								Value: aws.String("prod"),
							},
						},
					},
					{
						InstanceId: aws.String("i-789"),
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("id_baz"),
								Value: aws.String("987"),
							},
							{
								Key:   aws.String("env"),
								Value: aws.String("qa"),
							},
						},
					},
				},
			},
		},
	}

	return output, nil
}
