package cmd

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"

	"github.com/disneystreaming/ssm-helpers/aws/session"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	ssmx "github.com/disneystreaming/ssm-helpers/ssm"
	"github.com/disneystreaming/ssm-helpers/ssm/invocation"
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

	addBaseFlags(cmd)
	addRunFlags(cmd)

	return cmd
}

func runCommand(cmd *cobra.Command, args []string) {
	var err error
	var instanceList, commandList, profileList, regionList []string
	var maxConcurrency, maxErrors string
	var targets []*ssm.Target

	// Get all of our CLI flag values
	if err = cmdutil.ValidateArgs(cmd, args); err != nil {
		log.Fatal(err)
	}

	if instanceList, err = cmdutil.GetFlagStringSlice(cmd, "instance"); err != nil {
		log.Fatal(err)
	}
	if commandList, err = getCommandList(cmd); err != nil {
		log.Fatal(err)
	}
	if targets, err = getTargetList(cmd); err != nil {
		log.Fatal(err)
	}

	if err := validateRunFlags(cmd, instanceList, commandList, targets); err != nil {
		log.Fatal(err)
	}

	if profileList, err = getProfileList(cmd); err != nil {
		log.Fatal(err)
	}
	if regionList, err = getRegionList(cmd); err != nil {
		log.Fatal(err)
	}

	if maxConcurrency, err = getMaxConcurrency(cmd); err != nil {
		log.Fatal(err)
	}
	if maxErrors, err = getMaxErrors(cmd); err != nil {
		log.Fatal(err)
	}

	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

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
		MaxConcurrency: aws.String(maxConcurrency),
		MaxErrors:      aws.String(maxErrors),
	}

	wg, output := sync.WaitGroup{}, invocation.ResultSafe{}

	// Set up our AWS session for each permutation of profile + region and iterate over them
	sessionPool := session.NewPool(profileList, regionList, log)
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
