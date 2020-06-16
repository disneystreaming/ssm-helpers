package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
)

func TestAppendSSMFilter(t *testing.T) {
	assert := assert.New(t)

	// Create our filter slice
	filters := []*ssm.InstanceInformationStringFilter{
		{
			Key:    aws.String("PingStatus"),
			Values: aws.StringSlice([]string{"Online"}),
		},
	}

	// Create the filter to be appended
	secondFilter := &ssm.InstanceInformationStringFilter{
		Key:    aws.String("InstanceIds"),
		Values: aws.StringSlice([]string{"i-123", "i-456"}),
	}

	AppendSSMFilter(&filters, secondFilter)

	assert.Lenf(filters, 2, "Function returned slice of size %d, expected a size of 2", len(filters))
}

func TestBuildFilters(t *testing.T) {
	assert := assert.New(t)

	// Create our filter slice
	filters := []*ssm.InstanceInformationStringFilter{
		{
			Key:    aws.String("PingStatus"),
			Values: aws.StringSlice([]string{"Online"}),
		},
	}

	// Create the tags to be built and appended to the filter
	tags := map[string]string{
		"env": "dev",
		"foo": "bar",
	}

	BuildFilters(tags, &filters)

	assert.Lenf(filters, 3, "Filter slice was built incorrectly, got %d filters, expected 3", len(filters))
}

func TestNewSSMInstanceFilter(t *testing.T) {
	assert := assert.New(t)

	filter := NewSSMInstanceFilter("InstanceIds", CommaSlice{"i-123", "i-456"})

	assert.Equalf(*filter.Key, "InstanceIds", "Wrong filter key added to filter, got %s, expected 'InstanceIds'", *filter.Key)
	assert.Truef(
		aws.StringValueSlice(filter.Values)[0] == "i-123" && aws.StringValueSlice(filter.Values)[1] == "i-456",
		"Incorrect instance values added to filter, got %s, expected 'i-123' and 'i-456'",
		aws.StringValueSlice(filter.Values),
	)
}

func TestNewSSMFilter(t *testing.T) {
	assert := assert.New(t)

	filter := newSSMFilter("filterkey", "foo", "bar", "baz")

	assert.Equalf(*filter.Key, "filterkey", "Method set filter key to incorrect value, got %s, expected 'filterkey'", *filter.Key)

	values := aws.StringValueSlice(filter.Values)

	assert.Lenf(values, 3, "Incorrect number of values added to filter, got %d, expected 3", len(aws.StringValueSlice(filter.Values)))
	assert.Truef(
		values[0] == "foo" && values[1] == "bar" && values[2] == "baz",
		"Incorrect values added to filter, got %s, expected 'foo bar baz'",
		values,
	)
}
