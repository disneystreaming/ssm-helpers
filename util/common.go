package util

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// CommaSplit is a function used to split a comma-delimited list of strings into a slice of strings
func CommaSplit(value string) (args []string) {
	args = strings.Split(value, ",")
	return args
}

// SemiSplit is a function used to split a semicolon-delimited list of strings into a slice of strings
func SemiSplit(value string) (args []string) {
	args = strings.Split(value, ";")
	return args
}

// SliceToMap takes a slice of key=value elements and modifies a map to add those elements
func SliceToMap(kvslice []string, filterMap *map[string]string) {
	var elements []string
	for i := 0; i < len(kvslice); i++ {
		elements = strings.Split(kvslice[i], "=")
		(*filterMap)[elements[0]] = elements[1]
	}
}

func SliceToTargets(kvslice []string) (targets []*ssm.Target) {
	var elements []string

	for i := 0; i < len(kvslice); i++ {
		elements = strings.Split(kvslice[i], "=")
		targets = append(targets, &ssm.Target{
			Key:    aws.String(elements[0]),
			Values: aws.StringSlice([]string{elements[1]}),
		})
	}

	return targets
}
