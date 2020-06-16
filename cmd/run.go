package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	awsx "github.com/disneystreaming/ssm-helpers/aws"
	"github.com/disneystreaming/ssm-helpers/aws/session"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	ssmx "github.com/disneystreaming/ssm-helpers/ssm"
	"github.com/disneystreaming/ssm-helpers/ssm/invocation"
	"github.com/disneystreaming/ssm-helpers/util"
)

var version = "devel"
var commit = "notpassed"

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
	cmdutil.AddLimitFlag(cmd, "Set a limit for the number of instance results returned per profile/region combination (0 = no limit)")
	return cmd
}

func runCommand(cmd *cobra.Command, args []string) {
	cmdutil.ValidateArgs(cmd, args)
	commandList := cmdutil.GetCommandFlagStringSlice(cmd)
	verboseFlag := cmdutil.GetFlagInt(cmd.Parent(), "verbose")
	dryRunFlag := cmdutil.GetFlagBool(cmd.Parent(), "dry-run")
	profileList := cmdutil.GetFlagStringSlice(cmd.Parent(), "profile")
	regionList := cmdutil.GetFlagStringSlice(cmd.Parent(), "region")
	filterList := cmdutil.GetFlagStringSlice(cmd.Parent(), "filter")
	limitFlag := cmdutil.GetFlagInt(cmd, "limit")
	instanceList := cmdutil.GetFlagStringSlice(cmd, "instance")
	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set logs to go to stdout by default
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		// Disable level truncation, timestamp, and pad out the level text to even it up
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	if cmdutil.GetFlagBool(cmd, "version") {
		fmt.Printf("Version: %s\tGit Commit Hash: %s\n", version, commit)
		os.Exit(0)
	}

	if verboseFlag == 0 && dryRunFlag {
		verboseFlag = 1
	}

	if verboseFlag == 3 {
		log.SetLevel(log.DebugLevel)
	}

	// Split our commands into an array of individual commands, if necessary

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

	if verboseFlag > 0 {
		log.Info("Command(s) to be executed: ", strings.Join(commandList, ","))
	}

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

			if verboseFlag > 0 && (len(info) > 0) {
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

			err = ssmx.RunInvocations(sess, svc, info[:limitFlag], params, dryRunFlag, completedInvocations)
			if err != nil {
				log.Error(err)
			}
		}(sess, &completedInvocations)
	}

	wg.Wait()

	// Hide results if --verbose is set to quiet or terse
	if verboseFlag > 1 && !dryRunFlag {
		log.Infof("%-24s %-15s %-15s %s\n", "Instance ID", "Region", "Profile", "Status")
	}

	var successCounter int
	var failedCounter int

	for _, v := range completedInvocations.InvocationResults {

		// Hide results if --verbose is set to quiet or terse
		if v.Status != "Success" {
			if verboseFlag > 1 {
				log.Errorf("%-24s %-15s %-15s %s", *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			}

			// Always output error info to stderr
			log.Error(*v.InvocationResult.StandardErrorContent)
			failedCounter++
		} else {
			if verboseFlag > 1 {
				log.Infof("%-24s %-15s %-15s %s", *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			}
			if verboseFlag > 2 {
				log.Info(*v.InvocationResult.StandardOutputContent)
			}
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
