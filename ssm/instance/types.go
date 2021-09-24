package instance

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/disneystreaming/ssm-helpers/util"
)

// InstanceInfoSafe allows for concurrent-safe access to a map of InstanceInfo data
type InstanceInfoSafe struct {
	sync.Mutex
	AllInstances map[string]InstanceInfo
}

// InstanceInfo is used to store information, including EC2 tags, about a particular instance
type InstanceInfo struct {
	InstanceID string
	Region     string
	Profile    string
	VpcId      string
	Tags       map[string]string
}

// FormatStringSlice is used to return a strings preformatted to the correct width for selection prompts
func (i *InstanceInfoSafe) FormatStringSlice(includeFields ...string) (outSlice []string) {
	stringBuffer := new(bytes.Buffer)

	// Set up our tabwriter to nicely space our output
	tw := tabwriter.NewWriter(stringBuffer, 5, 4, 2, ' ', 0)

	// Set up our header string
	headerString := "Instance ID\tRegion\tProfile\t"
	for _, fieldName := range includeFields {
		headerString = fmt.Sprintf("%s%s\t", headerString, fieldName)
	}
	fmt.Fprintf(tw, "%s\n", headerString)

	for _, v := range i.AllInstances {
		fmt.Fprintf(tw, "%v\n", v.FormatString(includeFields...))
	}

	tw.Flush()

	return strings.Split(stringBuffer.String(), "\n")
}

// FormatString returns a string with various information about a given instance
func (i *InstanceInfo) FormatString(includeFields ...string) string {
	// Formatted string will always contain at least base info
	formattedString := fmt.Sprintf("%s\t%s\t%s\t", i.InstanceID, i.Region, i.Profile)
	fields := i.Tags
	fields["VpcId"] = i.VpcId

	for _, v := range includeFields {
		formattedString = fmt.Sprintf("%s%s\t", formattedString, fields[v])
	}

	return formattedString
}

// InstanceSlice is the type of object passed to flag.Var() that holds AWS instance IDs
type InstanceSlice []string

// String() returns the string representation of an *InstanceSlice object
func (i *InstanceSlice) String() string {
	return fmt.Sprintf("%s", *i)
}

// Set splits the provided comma-delimited list of instance IDs into an InstanceSlice, then sets the value of the caller to the new slice
func (i *InstanceSlice) Set(value string) error {
	list := util.CommaSplit(value)
	for _, v := range list {
		*i = append(*i, v)
	}
	return nil
}
