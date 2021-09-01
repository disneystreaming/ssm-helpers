package resolver

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockedEC2 struct {
	ec2iface.EC2API
	DescribeNetworkInterfacesOutput []*ec2.DescribeNetworkInterfacesOutput
}

func (c *mockedEC2) DescribeNetworkInterfacesPages(input *ec2.DescribeNetworkInterfacesInput, fn func(*ec2.DescribeNetworkInterfacesOutput, bool) bool) error {
	totalPages := len(c.DescribeNetworkInterfacesOutput)
	for i, output := range c.DescribeNetworkInterfacesOutput {
		isLastPage := (i == (totalPages - 1))
		if breakLoop := fn(output, isLastPage); breakLoop {
			break
		}
	}
	return nil
}

var (
	exampleInstanceId                             = "i-1234567890abcdef0"
	examplePrivateIpAddress                       = "127.0.0.1"
	singleResponseDescribeNetworkInterfacesOutput = ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: []*ec2.NetworkInterface{
			{
				Attachment: &ec2.NetworkInterfaceAttachment{
					InstanceId: &exampleInstanceId,
				},
				PrivateIpAddress: &examplePrivateIpAddress,
			},
		},
		NextToken: nil,
	}
)

func TestHostnameResolver(t *testing.T) {
	assert := assert.New(t)

	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)

	t.Run("test passed ip causes a short circuit", func(t *testing.T) {
		validIp, err := resolveHostnameToFirst(examplePrivateIpAddress)
		assert.Nil(err)
		assert.NotNil(validIp)
		assert.EqualValues(validIp, net.ParseIP(examplePrivateIpAddress))
	})

	t.Run("test invalid ip throws an error", func(t *testing.T) {
		validIp, err := resolveHostnameToFirst("127.1.4")
		assert.NotNil(err)
		assert.Nil(validIp)
	})

	t.Run("test passed passed hostname resolves to an IP", func(t *testing.T) {
		validIp, err := resolveHostnameToFirst("example.com")
		assert.Nil(err)
		assert.NotNil(validIp)
	})

	t.Run("test mock resolver returns a valid instance id", func(t *testing.T) {
		mockClient := &mockedEC2{DescribeNetworkInterfacesOutput: []*ec2.DescribeNetworkInterfacesOutput{&singleResponseDescribeNetworkInterfacesOutput}}
		testResolver := NewHostnameResolver([]string{examplePrivateIpAddress})

		resp, err := testResolver.ResolveToInstanceId(mockClient)
		assert.Nil(err)
		assert.EqualValues(resp, []string{exampleInstanceId})
	})
}
