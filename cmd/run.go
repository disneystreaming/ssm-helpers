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
	return cmd
}

func runCommand(cmd *cobra.Command, args []string) {
	cmdutil.ValidateArgs(cmd, args)

	commandList := cmdutil.GetCommandFlagStringSlice(cmd)
	profileList := cmdutil.GetFlagStringSlice(cmd.Parent(), "profile")
	regionList := cmdutil.GetFlagStringSlice(cmd.Parent(), "region")
	filterList := cmdutil.GetFlagStringSlice(cmd.Parent(), "filter")
	instanceList := cmdutil.GetFlagStringSlice(cmd, "instance")
	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(instanceList) > 0 && len(filterList) > 0 {
		cmdutil.UsageError(cmd, "The --filter and --instance flags cannot be used simultaneously.")
		os.Exit(1)
	}

	if len(instanceList) > 50 {
		cmdutil.UsageError(cmd, "The --instance flag can only be used to specify a maximum of 50 instances.")
	}

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

	// Set up our AWS session for each permutation of profile + region
	sessionPool := session.NewPoolSafe(profileList, regionList)

	// Convert the filter slice to a map
	targets := []*ssm.Target{}

	if len(filterList) > 0 {
		targets = util.SliceToTargets(filterList)
	}

	log.Info("Command(s) to be executed:\n", strings.Join(commandList, "\n"))

	sciInput := &ssm.SendCommandInput{
		InstanceIds:  aws.StringSlice(instanceList),
		Targets:      targets,
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{
			/*
				ssm.SendCommandInput objects require parameters for the DocumentName chosen

				For AWS-RunShellScript, the only required parameter is "commands",
				which is the shell command to be executed on the target. To emulate
				the original script, we also set "executionTimeout" to 10 minutes.
			*/
			"commands":         aws.StringSlice(commandList),
			"executionTimeout": aws.StringSlice([]string{"600"}),
		},
	}

	// Set up our AWS session for each permutation of profile + region
	sessionPool := session.NewPoolSafe(profileList, regionList, log)
	wg, output := sync.WaitGroup{}, invocation.ResultSafe{}

	for _, sess := range sessionPool.Sessions {
		wg.Add(1)
		ssmClient := ssm.New(sess.Session)
		log.Debugf("Starting invocation targeting account %s in %s", sess.ProfileName, *sess.Session.Config.Region)
		go ssmx.RunInvocations(sess, ssmClient, &wg, sciInput, &output)
	}

	wg.Wait() // Wait for each account/region combo to finish

	resultFormat := "%-24s %-15s %-15s %s"
	var successCounter, failedCounter int

	// Output our results
	log.Infof(resultFormat, "Instance ID", "Region", "Profile", "Status")
	for _, v := range output.InvocationResults {
		switch v.Status {
		case "Success":
			log.Infof(resultFormat, *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			successCounter++
		default:
			log.Errorf(resultFormat, *v.InvocationResult.InstanceId, v.Region, v.ProfileName, *v.InvocationResult.StatusDetails)
			failedCounter++
		}

		// stdout is always written back at info level
		if *v.InvocationResult.StandardOutputContent != "" {
			log.Info(*v.InvocationResult.StandardOutputContent)
		}

		// stderr is written back at warn if the invocation was successful, and error if not
		if *v.InvocationResult.StandardErrorContent != "" {
			if v.Status == "Success" {
				log.Warn(*v.InvocationResult.StandardErrorContent)
			} else {
				log.Error(*v.InvocationResult.StandardErrorContent)
			}
		}
	}

	log.Infof("Execution results: %d SUCCESS, %d FAILED", successCounter, failedCounter)
	if failedCounter > 0 { // Exit code 1 to indicate that there was some sort of error returned from invocation
		os.Exit(1)
	}

	return
}
