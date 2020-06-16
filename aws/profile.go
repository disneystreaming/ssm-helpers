package aws

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/mitchellh/go-homedir"
)

// GetAWSProfiles accesses a user's ~/.aws/config file and uses regex matching to extract the
// list of profile names for use with the --all-profiles flag. The path to any arbitrary config
// file can also be provided as an argument.
func GetAWSProfiles(profilePath ...string) (profiles []string, err error) {
	var configFile *os.File
	// Try and open our .aws/config file to enumerate profiles
	if profilePath == nil {
		dir, _ := homedir.Dir()
		configFile, err = os.Open(dir + "/.aws/config")
		if err != nil {
			return nil, err
		}
	} else {
		if len(profilePath) > 1 {
			return nil, fmt.Errorf("Multiple profile path arguments provided, please check your syntax: %s", profilePath)
		}
		_, err := os.Stat(profilePath[0])
		if os.IsNotExist(err) {
			return nil, err
		}
		configFile, err = os.Open(profilePath[0])
		if err != nil {
			return nil, err
		}
	}

	reg := regexp.MustCompile(`\[profile (.*)\]`)
	// Parse through config file to get a list of all profiles
	var data []string
	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		// FindStringSubmatch returns the entire match, including "[profile ]", match[1] contains the first capture
		match := reg.FindStringSubmatch(scanner.Text())
		if match != nil {
			data = append(data, match[1])
		}
	}

	return data, err
}
