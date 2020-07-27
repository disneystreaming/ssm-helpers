package cmd

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"

	awsx "github.com/disneystreaming/ssm-helpers/aws"
	"github.com/disneystreaming/ssm-helpers/aws/session"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	ssmx "github.com/disneystreaming/ssm-helpers/ssm"
	"github.com/disneystreaming/ssm-helpers/ssm/invocation"
	"github.com/disneystreaming/ssm-helpers/util"
)

func newCommandSSMRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "execute commands using the AWS-RunShellScript document",
		Long:  "foo bar baz",
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(cmd, args)
		},
	}

	cmdutil.AddCommandFlag(cmd)
	cmdutil.AddFileFlag(cmd, "Specify the path to a shell script to use as input for the AWS-RunShellScript document.\nThis can be used in combination with the --commands/-c flag, and will be run after the specified commands.")
	cmdutil.AddLimitFlag(cmd, 0, "Set a limit for the number of instance results returned per profile/region combination (0 = no limit)")
	cmdutil.AddMaxConcurrencyFlag(cmd, "50", "Max targets to run the command in parallel. Both numbers, such as 50, and percentages, such as 50%, are allowed")
	cmdutil.AddMaxErrorsFlag(cmd, "0", "Max errors allowed before running on additional targets. Both numbers, such as 10, and percentages, such as 10%, are allowed")
	return cmd
}

func runCommand(cmd *cobra.Command, args []string) {
	cmdutil.ValidateArgs(cmd, args)

	commandList := cmdutil.GetCommandFlagStringSlice(cmd)
	dryRunFlag := cmdutil.GetFlagBool(cmd.Parent(), "dry-run")
	profileList := cmdutil.GetFlagStringSlice(cmd.Parent(), "profile")
	regionList := cmdutil.GetFlagStringSlice(cmd.Parent(), "region")
	filterList := cmdutil.GetFlagStringSlice(cmd.Parent(), "filter")
	limitFlag := cmdutil.GetFlagInt(cmd, "limit")
	maxConcurrencyFlag := cmdutil.GetFlagString(cmd, "max-concurrency")
	maxErrorsFlag := cmdutil.GetFlagString(cmd, "max-errors")
	instanceList := cmdutil.GetFlagStringSlice(cmd, "instance")
	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

	// If the --commands and --file options are specified, we append the script contents to the specified commands
	if inputFile := cmdutil.GetFlagString(cmd, "file"); inputFile != "" {
		// Open our file for reading
		file, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("Could not open file at %s\n%s", inputFile, err)
		}

		defer file.Close()

		// Grab each line of the script and append it to the command slice
		// Scripts using a line continuation character (\) will work fine here too!
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			commandList = append(commandList, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Issue when trying to read input file\n%s", err)
		}
	}

	if commandList == nil || len(commandList) == 0 {
		cmdutil.UsageError(cmd, "Please specify a command to be run on your instances.")
		os.Exit(1)
	}

	// ssm.SendCommandInput objects require parameters for the DocumentName chosen
	params := &invocation.RunShellScriptParameters{
		/*
			For AWS-RunShellScript, the only required parameter is "commands",
			which is the shell command to be executed on the target. To emulate
			the original script, we also set "executionTimeout" to 10 minutes.
		*/
		"commands":         aws.StringSlice(commandList),
		"executionTimeout": aws.StringSlice([]string{"600"}),
	}

	log.Info("Command(s) to be executed: ", strings.Join(commandList, ","))

	if len(profileList) == 0 {
		env, exists := os.LookupEnv("AWS_PROFILE")
		if exists {
			profileList = []string{env}
		} else {
			profileList = []string{"default"}
		}
	}

	if len(regionList) == 0 {
		env, exists := os.LookupEnv("AWS_REGION")
		if exists == false {
			regionList = []string{env}
		}
	}

	// If --all-profiles is set, we call getAWSProfiles() and iterate through the user's ~/.aws/config
	if allProfilesFlag := cmdutil.GetFlagBool(cmd, "all-profiles"); allProfilesFlag {
		profileList, err := awsx.GetAWSProfiles()
		if profileList == nil || err != nil {
			log.Error("Could not load profiles.", err)
			os.Exit(1)
		}
	}
	if !cmdutil.ValidateIntOrPercantageValue(maxConcurrencyFlag) {
		log.Error(`--max-concurrency: Invalid value passed
			Length Constraints: Minimum length of 1. Maximum length of 7.
			Pattern: ^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`)
		os.Exit(1)
	}

	if !cmdutil.ValidateIntOrPercantageValue(maxErrorsFlag) {
		log.Error(`--max-errors: Invalid value passed
			Length Constraints: Minimum length of 1. Maximum length of 7.
			Pattern: ^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`)
		os.Exit(1)
	}

	// Set up our AWS session for each permutation of profile + region
	sessionPool := session.NewPoolSafe(profileList, regionList)

	// Set up our filters
	var filterMaps []map[string]string

	// Convert the filter slice to a map
	filterMap := make(map[string]string)

	if len(filterList) > 0 {
		util.SliceToMap(filterList, &filterMap)
		filterMaps = append(filterMaps, filterMap)
	}

	var completedInvocations invocation.ResultSafe
	var wg sync.WaitGroup

	for _, sess := range sessionPool.Sessions {
		wg.Add(1)
		go func(sess *session.Pool, completedInvocations *invocation.ResultSafe) {
			defer wg.Done()
			instanceChan := make(chan []*ssm.InstanceInformation)
			errChan := make(chan error)
			svc := ssm.New(sess.Session)

			go ssmx.GetInstanceList(svc, filterMaps, instanceList, false, instanceChan, errChan)
			info, err := <-instanceChan, <-errChan

			if err != nil {
				log.Debugf("AWS Session Parameters: %s, %s", *sess.Session.Config.Region, sess.ProfileName)
				log.Error(err)
			}

			if len(info) == 0 {
				return
			}

			if len(info) > 0 {
				log.Infof("Fetched %d instances for account [%s] in [%s].", len(info), sess.ProfileName, *sess.Session.Config.Region)
				if dryRunFlag {
					log.Info("Targeted instances:")
					for _, instance := range info {
						log.Infof("%s", *instance.InstanceId)
					}
				}
			}

			if limitFlag == 0 || limitFlag > len(info) {
				limitFlag = len(info)
			}

			if err = ssmx.RunInvocations(sess, svc, info[:limitFlag], params, dryRunFlag, maxConcurrencyFlag, maxErrorsFlag, completedInvocations); err != nil {
				log.Error(err)
			}
		}(sess, &completedInvocations)
	}

	wg.Wait()

	// Hide results if --verbose is set to quiet or terse
	if !dryRunFlag {
		log.Infof("%-24s %-15s %-15s %s\n", "Instance ID", "Region", "Profile", "Status")
	}

	var successCounter int
	var failedCounter int

	for _, v := range completedInvocations.InvocationResults {

		// Hide results if --verbose is set to quiet or terse
		if v.Status != "Success" {
			// Always output error info to stderr
			log.Errorf("%-24s %-15s %-15s %s", *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			log.Error(*v.InvocationResult.StandardErrorContent)

			failedCounter++
		} else {
			// Output stdout from invocations to stdout
			log.Infof("%-24s %-15s %-15s %s", *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			log.Info(*v.InvocationResult.StandardOutputContent)

			successCounter++
		}

	}

	if !dryRunFlag {
		log.Infof("Execution results: %d SUCCESS, %d FAILED", successCounter, failedCounter)
		if failedCounter > 0 {
			// Exit code 1 to indicate that there was some sort of error returned from invocation
			os.Exit(1)
		}
	}

	return
}
