package ssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// AppendSSMFilter appends a single *ssm.InstanceInformationStringFilter to an existing array of filters, then returns the appended array
func AppendSSMFilter(filters *[]*ssm.InstanceInformationStringFilter, filterToAdd *ssm.InstanceInformationStringFilter) {
	*filters = append(*filters, filterToAdd)
}

// BuildFilters takes a map of tags (in key[value] format) and converts them into a format that is compatible
// with *ssm.InstanceInformationStringFilter, updating a provided []*ssm.InstanceInformationStringFilter object
func BuildFilters(commonTags map[string]string, filtersList *[]*ssm.InstanceInformationStringFilter) {

	merged := make(map[string]string)

	for k, v := range commonTags {
		merged[k] = v
	}

	for k, v := range merged {
		filter := newSSMFilter("tag:"+k, v)
		*filtersList = append(*filtersList, filter)
	}

	return
}

// NewSSMInstanceFilter takes a name and any number of instance IDs to create an *ssm.InstanceInformationStringFilter
// that can be used to select specific SSM-managed instances instead of solely searching based on tags
func NewSSMInstanceFilter(name string, values CommaSlice) *ssm.InstanceInformationStringFilter {
	stringPointerSlice := aws.StringSlice(values)
	filter := &ssm.InstanceInformationStringFilter{
		Key:    aws.String(name),
		Values: stringPointerSlice,
	}
	return filter
}

// newSSMFilter takes a name and any number of string values and returns an *ssm.InstanceInformationStringFilter object,
// as the values must be inside of a []*string object to be used in the filter. Yes, it's weird.
func newSSMFilter(name string, values ...string) *ssm.InstanceInformationStringFilter {
	awsValues := []*string{}
	for _, value := range values {
		awsValues = append(awsValues, aws.String(value))
	}

	filter := &ssm.InstanceInformationStringFilter{
		Key:    aws.String(name),
		Values: awsValues,
	}
	return filter
}
