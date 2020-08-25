package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	awsx "github.com/disneystreaming/ssm-helpers/aws"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	"github.com/disneystreaming/ssm-helpers/cmd/logutil"
	"github.com/disneystreaming/ssm-helpers/util"
)

func addBaseFlags(cmd *cobra.Command) {
	cmdutil.AddAllProfilesFlag(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmdutil.AddFilterFlag(cmd)
	cmdutil.AddInstanceFlag(cmd)
	cmdutil.AddProfileFlag(cmd)
	cmdutil.AddRegionFlag(cmd)
}

func addRunFlags(cmd *cobra.Command) {
	cmdutil.AddCommandFlag(cmd)
	cmdutil.AddFileFlag(cmd, "Specify the path to a shell script to use as input for the AWS-RunShellScript document.\nThis can be used in combination with the --commands/-c flag, and will be run after the specified commands.")
	cmdutil.AddMaxConcurrencyFlag(cmd, "50", "Max targets to run the command in parallel. Both numbers, such as 50, and percentages, such as 50%, are allowed")
	cmdutil.AddMaxErrorsFlag(cmd, "0", "Max errors allowed before running on additional targets. Both numbers, such as 10, and percentages, such as 10%, are allowed")
}

func addSessionFlags(cmd *cobra.Command) {
	cmdutil.AddTagFlag(cmd)
	cmdutil.AddSessionNameFlag(cmd, "ssm-session")
	cmdutil.AddLimitFlag(cmd, 10, "Set a limit for the number of instance results returned per profile/region combination.")
}

func getCommandList(cmd *cobra.Command) (commandList []string, err error) {
	if commandList, err = cmdutil.GetCommandFlagStringSlice(cmd); err != nil {
		return nil, err
	}

	// If the --commands and --file options are specified, we append the script contents to the specified commands
	if inputFile, err := cmdutil.GetFlagString(cmd, "file"); inputFile != "" && err == nil {
		if err = util.ReadScriptFile(inputFile, &commandList); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return commandList, nil
}

func getRegionList(cmd *cobra.Command) ([]string, error) {
	regions, err := cmdutil.GetFlagStringSlice(cmd, "region")
	if err != nil {
		return []string{}, err
	}

	if len(regions) != 0 {
		return regions, nil
	}

	// AWS_REGION is the deprecated env variable used to configure the region
	region, exists := os.LookupEnv("AWS_REGION")
	if exists {
		return []string{region}, nil
	}

	defaultRegion, exists := os.LookupEnv("AWS_DEFAULT_REGION")
	if exists {
		return []string{defaultRegion}, nil
	}

	return []string{""}, nil
}

func getTargetList(cmd *cobra.Command) (targets []*ssm.Target, err error) {
	var filterList []string
	if filterList, err = cmdutil.GetFlagStringSlice(cmd, "filter"); err != nil {
		return nil, err
	}

	return util.SliceToTargets(filterList), nil
}

func getProfileList(cmd *cobra.Command) (profileList []string, err error) {
	if profileList, err = cmdutil.GetFlagStringSlice(cmd, "profile"); err != nil {
		return nil, err
	}

	var allProfilesFlag bool
	if allProfilesFlag, err = cmdutil.GetFlagBool(cmd, "all-profiles"); err != nil {
		return nil, err
	}

	if len(profileList) > 0 && allProfilesFlag {
		return nil, cmdutil.UsageError(cmd, "The --profile and --all-profiles flags cannot be used simultaneously.")
	}

	if allProfilesFlag { // If --all-profiles is set, we call getAWSProfiles() and iterate through the user's ~/.aws/config
		if profileList, err = awsx.GetAWSProfiles(); profileList == nil || err != nil {
			return nil, fmt.Errorf("Could not load profiles.\n%v", err)
		}
	}

	if len(profileList) == 0 {
		if env, exists := os.LookupEnv("AWS_PROFILE"); exists {
			profileList = []string{env}
		} else {
			profileList = []string{"default"}
		}
	}

	return profileList, nil
}

func getMaxConcurrency(cmd *cobra.Command) (maxConcurrency string, err error) {
	if maxConcurrency, err = cmdutil.GetFlagString(cmd, "max-concurrency"); err != nil {
		return "", err
	}

	if !cmdutil.ValidateMaxConcurrency(maxConcurrency) {
		return "", fmt.Errorf(`--max-concurrency: Invalid value passed
Length Constraints: Minimum length of %d. Maximum length of %d.
Pattern: %q`, 1, 7, "^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$")
	}

	return maxConcurrency, nil
}

func getMaxErrors(cmd *cobra.Command) (maxErrors string, err error) {
	if maxErrors, err = cmdutil.GetFlagString(cmd, "max-errors"); err != nil {
		return "", err
	}

	if !cmdutil.ValidateMaxErrors(maxErrors) {
		return "", fmt.Errorf(`--max-errors: Invalid value passed
Length Constraints: Minimum length of %d. Maximum length of %d.
Pattern: %q`, 1, 7, "^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$")
	}

	return maxErrors, nil
}

func validateSessionFlags(cmd *cobra.Command, instanceList []string, filterList map[string]string) error {
	if len(instanceList) > 0 && len(filterList) > 0 {
		return cmdutil.UsageError(cmd, "The --filter and --instance flags cannot be used simultaneously.")
	}

	return nil
}

// validateRunFlags validates the usage of certain flags required by the run subcommand
func validateRunFlags(cmd *cobra.Command, instanceList []string, commandList []string, filterList []*ssm.Target) error {
	if len(instanceList) > 0 && len(filterList) > 0 {
		return cmdutil.UsageError(cmd, "The --filter and --instance flags cannot be used simultaneously.")
	}

	if len(instanceList) == 0 && len(filterList) == 0 {
		return cmdutil.UsageError(cmd, "You must supply target arguments using either the --filter or --instance flags.")
	}

	if len(instanceList) > 50 {
		return cmdutil.UsageError(cmd, "The --instance flag can only be used to specify a maximum of 50 instances.")
	}

	if len(commandList) == 0 {
		return cmdutil.UsageError(cmd, "Please specify a command to be run on your instances.")
	}

	return nil
}

func setLogLevel(cmd *cobra.Command, log *logrus.Logger) (err error) {
	v, err := cmdutil.GetFlagInt(cmd, "verbose")
	if err != nil {
		return err
	}

	log.SetLevel(logutil.IntToLogLevel(v))
	return nil
}
