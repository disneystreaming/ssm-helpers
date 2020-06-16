package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	"github.com/disneystreaming/ssm-helpers/ec2"
	"github.com/disneystreaming/ssm-helpers/ssm/instance"
)

func TestCreateSSMDescribeInstanceInput(t *testing.T) {
	assert := assert.New(t)

	var filterSlice []map[string]string
	filters := map[string]string{
		"foo": "1",
		"bar": "2",
		"baz": "3",
	}

	moreFilters := map[string]string{
		"fizz": "4",
		"buzz": "5",
	}
	filterSlice = append(filterSlice, filters, moreFilters)

	instances := &CommaSlice{"i-12345", "i-67890"}

	instanceInput := CreateSSMDescribeInstanceInput(filterSlice, *instances)

	// Ensure that appending of filters is working correctly for multiple tags
	assert.Lenf(instanceInput.Filters, 6, "Filter slice has wrong number of entries, got %d, expected 6 items (instance IDs + all tags)", len(instanceInput.Filters))

	// Ensure that MaxResults is set to 50, which is the maximum number of results returned per page
	assert.Equalf(*instanceInput.MaxResults, int64(50), "MaxResults was not set to the correct value, got %d, expected 50", *instanceInput.MaxResults)
}

func TestAddInstanceInfo(t *testing.T) {
	assert := assert.New(t)

	// Create our tag data
	tagStruct := []ec2.InstanceTags{
		{
			Tags: map[string]string{
				"foo": "1",
				"bar": "2",
			},
			InstanceID: "i-123",
		},
	}

	// Initialize our mutex-safe map
	ip := &instance.InstanceInfoSafe{
		AllInstances: map[string]instance.InstanceInfo{},
	}

	addInstanceInfo(aws.String("i-123"), tagStruct, ip, "testprofile", "us-east-1")

	assert.Equalf(
		ip.AllInstances["i-123"].InstanceID, "i-123",
		"Instance info object returned wrong instance ID, got %s, expected 'i-123'",
		ip.AllInstances["i-123"].InstanceID,
	)

	assert.Truef(
		ip.AllInstances["i-123"].Tags["foo"] == "1" && ip.AllInstances["i-123"].Tags["bar"] == "2",
		"Instance info object returned incorrect tag data, got %s",
		ip.AllInstances["i-123"].Tags,
	)

}
