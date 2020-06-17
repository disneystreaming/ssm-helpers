package cmdutil

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AddProfileFlag adds --profile to command
func AddProfileFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("profile", "p", nil, "Specify a specific profile to use with your API calls.\nMultiple allowed, delimited by commas (e.g. --profile profile1,profile2)")
}

// AddRegionFlag adds --region to command
func AddRegionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("region", "r", nil, "Specify a specific region to use with your API calls.\n"+
		"This option will override any profile settings in your config file.\n"+
		"Multiple allowed, delimited by commas (e.g. --region us-east-1,us-west-2)\n\n"+
		"[NOTE] Mixing --profile and --region will result in your command targeting every matching instance in the selected profiles and regions.\n"+
		"e.g., \"--profile foo,bar,baz --region us-east-1,us-west-2,eu-east-1\" will target instances in each of the profile/region combinations:\n"+
		"\t\"foo@us-east-1, foo@us-west-2, foo@eu-east-1\"\n"+
		"\t\"bar@us-east-1, bar@us-west-2, bar@eu-east-1\"\n"+
		"\t\"baz@us-east-1, baz@us-west-2, baz@eu-east-1\"\n"+
		"Please be careful.")
}

// AddFilterFlag adds --filter to command
func AddFilterFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("filter", "f", nil, "Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).\nMultiple allowed, delimited by commas (e.g. env=dev,foo=bar)")
}

// AddDryRunFlag adds --dry-run to command
func AddDryRunFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("dry-run", false, "Retrieve the list of profiles, regions, and instances your command(s) would target")
}

// AddVerboseFlag adds --verbose to command
func AddVerboseFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().IntP("verbose", "v", 2, "Sets verbosity of output:\n0 = quiet, 1 = terse, 2 = standard, 3 = debug")
}

// AddLimitFlag adds --limit to command
func AddLimitFlag(cmd *cobra.Command, limit int, desc string) {
	cmd.PersistentFlags().IntP("limit", "l", limit, desc)
}

// AddInstanceFlag adds --instance to command
func AddInstanceFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("instance", "i", nil, "Specify what instance IDs you want to target.\nMultiple allowed, delimited by commas (e.g. --instance i-12345,i-23456)")
}

// AddAllProfilesFlag adds --all-profiles to command
func AddAllProfilesFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("all-profiles", false, "[USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.")

}

// AddCommandFlag adds --command to command
func AddCommandFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("command", "c", "", "Specify any number of commands to be run.\nMultiple allowed, enclosed in double quotes and delimited by semicolons (e.g. --comands \"hostname; uname -a\")")
}

// AddFileFlag adds --file to command
func AddFileFlag(cmd *cobra.Command, desc string) {
	cmd.Flags().String("file", "", desc)
}

// AddTagFlag adds --tag to command
func AddTagFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("tag", "t", nil, "Adds the specified tag as an additional column to be displayed during the instance selection prompt.")
}

// AddSessionNameFlag adds --session-name to command
func AddSessionNameFlag(cmd *cobra.Command, defaultName string) {
	cmd.Flags().String("session-name", defaultName, "Specify a name for the tmux session created when multiple instances are selected")
}

// ValidateArgs makes sure nothing extra was passed on CLI
func ValidateArgs(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		UsageError(cmd, "Unexpected args: %v", strings.Join(args, " "))
	}
}

// UsageError Prints error and tells users to use -h
func UsageError(cmd *cobra.Command, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s\nSee '%s -h' for help and examples.\n", msg, cmd.CommandPath())
	os.Exit(1)
}

// GetCommandFlagStringSlice returns the []string value of a String() flag, delimited by semicolons
func GetCommandFlagStringSlice(cmd *cobra.Command) []string {
	commandString, err := cmd.Flags().GetString("command")
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"flag":    "command",
				"command": cmd.Name(),
			}).
			Error("could not fetch flag")
	}

	return readAsSSV(commandString)
}

// GetFlagStringSlice returns the []string value of a StringSlice() flag
func GetFlagStringSlice(cmd *cobra.Command, flag string) []string {
	s, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"flag":    flag,
				"command": cmd.Name(),
			}).
			Error("could not fetch flag")
	}
	return s
}

// GetFlagString returns the string value of a String() flag
func GetFlagString(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"flag":    flag,
				"command": cmd.Name(),
			}).
			Error("could not fetch flag")
	}
	return s
}

// GetFlagBool returns the bool value from a Bool() flag
func GetFlagBool(cmd *cobra.Command, flag string) bool {
	s, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"flag":    flag,
				"command": cmd.Name(),
			}).
			Error("could not fetch flag")
	}
	return s
}

// GetFlagInt returns the integer value from an Int() flag
func GetFlagInt(cmd *cobra.Command, flag string) int {
	s, err := cmd.Flags().GetInt(flag)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"flag":    flag,
				"command": cmd.Name(),
			}).
			Error("could not fetch flag")
	}
	return s
}

// GetMapFromStringSlice returns a k,v map from a StringSlice() flag
func GetMapFromStringSlice(cmd *cobra.Command, flag string) map[string]string {
	m := make(map[string]string)
	slice := GetFlagStringSlice(cmd, flag)
	squashSlice := squashParamsSlice(slice, cmd)

	for _, v := range squashSlice {
		if !strings.Contains(v, "=") {
			UsageError(cmd, "Invalid Parameter format: %s\n", v)
		}
		// Only split to retun a max of 2 values. This will take string
		// key=value= and return ["key", "value="]
		kv := strings.SplitN(v, "=", 2)
		m[kv[0]] = kv[1]
	}

	return m
}

// Reformats CSV params passed via CLI
// e.g. Input:	["Env=dev", "ElbSecurityGroups=sg-1234", "sg-5678", "App=grafana"]
// 		Output:	["Env=dev", "ElbSecurityGroups=sg-1234,sg-5678", "App=grafana"]
func squashParamsSlice(slice []string, cmd *cobra.Command) []string {
	sqS := make([]string, 0, len(slice))
	index := -1
	if len(slice) != 0 {
		for _, v := range slice {
			if !strings.Contains(v, "=") {
				if index < 0 {
					UsageError(cmd, "Invalid Parameter format:%s\n", v)
				}
				sqS[index] = sqS[index] + "," + v
			} else {
				sqS = append(sqS, v)
				index++
			}
		}
	}

	return sqS
}
