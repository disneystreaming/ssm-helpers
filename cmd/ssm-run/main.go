package main

import (
	"bufio"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	awshelpers "github.com/disneystreaming/go-ssmhelpers/aws"
	"github.com/disneystreaming/go-ssmhelpers/aws/session"
	ssmhelpers "github.com/disneystreaming/go-ssmhelpers/ssm"
	"github.com/disneystreaming/go-ssmhelpers/ssm/invocation"
	"github.com/disneystreaming/go-ssmhelpers/util"
)

var commandList []string
var myCommands ssmhelpers.SemiSlice
var myInstances ssmhelpers.CommaSlice
var myFilters ssmhelpers.CommaSlice
var myProfiles ssmhelpers.CommaSlice
var myRegions ssmhelpers.CommaSlice
var allProfilesFlag bool
var verboseFlag int
var dryRunFlag bool
var limitFlag int

func main() {
	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set logs to go to stdout by default
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		// Disable level truncation, timestamp, and pad out the level text to even it up
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	errLog := logrus.New()
	errLog.SetFormatter(&log.TextFormatter{
		// Disable level truncation, timestamp, and pad out the level text to even it up
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	// Flag for commands to be run
	flag.Var(&myCommands, "commands", "Specify any number of commands to be run.\nMultiple allowed, enclosed in double quotes and delimited by semicolons (e.g. --comands \"hostname; uname -a\")")
	flag.Var(&myCommands, "c", "--commands (shorthand)")

	// Flag to use a shell script as the document input
	inputFile := flag.String("file", "", "Specify the path to a shell script to use as input for the AWS-RunShellScript document.\nThis can be used in combination with the --commands/-c flag, and will be run after the specified commands.")

	// Flag to indicate that the user wants to perform a dry run
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Retrieve the list of profiles, regions, and instances your command(s) would target.")

	// Flag to indicate that the user wants to execute a command against all of their configured profiles
	flag.BoolVar(&allProfilesFlag, "all-profiles", false, "[USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.")

	// Flag to enable increasingly-verbose output
	flag.IntVar(&verboseFlag, "log-level", 0, "Sets verbosity of output:\n0 = quiet, 1 = terse, 2 = standard, 3 = debug")

	// Flag for instance selection
	flag.Var(&myInstances, "instances", "Specify what instance IDs you want to target.\nMultiple allowed, delimited by commas (e.g. --instances i-12345,i-23456)")
	flag.Var(&myInstances, "i", "--instances (shorthand)")

	// Flags for filters
	flag.Var(&myFilters, "filter", "Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).\nMultiple allowed, delimited by commas (e.g. env=dev,foo=bar)")
	flag.Var(&myFilters, "f", "--filter (shorthand)")

	// Flags for profiles/regions
	flag.Var(&myProfiles, "profiles", "Specify a specific profile to use with your API calls.\nMultiple allowed, delimited by commas (e.g. --profiles profile1,profile2)")
	flag.Var(&myProfiles, "p", "--profiles (shorthand)")
	flag.Var(&myRegions, "regions", "Specify a specific region to use with your API calls.\n"+
		"This option will override any profile settings in your config file.\n"+
		"Multiple allowed, delimited by commas (e.g. --regions us-east-1,us-west-2)\n\n"+
		"[NOTE] Mixing --profiles and --regions will result in your command targeting every matching instance in the selected profiles and regions.\n"+
		"e.g., \"--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1\" will target instances in each of the profile/region combinations:\n"+
		"\t\"foo@us-east-1, foo@us-west-2, foo@eu-east-1\"\n"+
		"\t\"bar@us-east-1, bar@us-west-2, bar@eu-east-1\"\n"+
		"\t\"baz@us-east-1, baz@us-west-2, baz@eu-east-1\"\n"+
		"Please be careful.")
	flag.Var(&myRegions, "r", "--regions (shorthand)")

	// Flag to set a limit to the number of instances returned by the SSM/EC2 API query
	flag.IntVar(&limitFlag, "limit", 0, "Set a limit for the number of instance results returned per profile/region combination (0 = no limit)")

	flag.Parse()

	if verboseFlag == 3 {
		log.SetLevel(log.DebugLevel)
	}

	// Split our commands into an array of individual commands, if necessary
	if myCommands == nil && *inputFile == "" {
		errLog.Error("Please specify a command to be run on your instances.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// If the --commands and --file options are specified, we append the script contents to the specified commands
	if *inputFile != "" {
		// Open our file for reading
		file, err := os.Open(*inputFile)
		if err != nil {
			log.Fatalf("Could not open file at %s\n%s", *inputFile, err)
		}

		defer file.Close()

		// Grab each line of the script and append it to the command slice
		// Scripts using a line continuation character (\) will work fine here too!
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			myCommands = append(myCommands, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Issue when trying to read input file\n%s", err)
		}
	}

	// ssm.SendCommandInput objects require parameters for the DocumentName chosen
	params := &invocation.RunShellScriptParameters{
		/*
			For AWS-RunShellScript, the only required parameter is "commands",
			which is the shell command to be executed on the target. To emulate
			the original script, we also set "executionTimeout" to 10 minutes.
		*/
		"commands":         aws.StringSlice([]string(myCommands)),
		"executionTimeout": aws.StringSlice([]string{"600"}),
	}

	if verboseFlag > 0 {
		log.Info("Command(s) to be executed: ", strings.Join(myCommands, ","))
	}

	if myProfiles == nil {
		env, exists := os.LookupEnv("AWS_PROFILE")
		if exists {
			myProfiles.Set(env)
		} else {
			myProfiles.Set("default")
		}
	}

	if myRegions == nil {
		env, exists := os.LookupEnv("AWS_REGION")
		if exists == false {
			myRegions.Set(env)
		}
	}

	// If --all-profiles is set, we call getAWSProfiles() and iterate through the user's ~/.aws/config
	if allProfilesFlag {
		profiles, err := awshelpers.GetAWSProfiles()
		if profiles != nil && err == nil {
			myProfiles = profiles
		} else {
			errLog.Error("Could not load profiles.", err)
			return
		}
	}

	// Set up our AWS session for each permutation of profile + region
	sessionPool := session.NewPoolSafe(myProfiles, myRegions)

	// Set up our filters
	var filterMaps []map[string]string

	// Convert the filter slice to a map
	filterMap := make(map[string]string)

	if len(myFilters) > 0 {
		util.SliceToMap(myFilters, &filterMap)
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

			go ssmhelpers.GetInstanceList(svc, filterMaps, myInstances, false, instanceChan, errChan)
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
			}

			if limitFlag == 0 || limitFlag > len(info) {
				limitFlag = len(info)
			}

			err = ssmhelpers.RunInvocations(sess, svc, info[:limitFlag], params, dryRunFlag, completedInvocations)
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
			errLog.Error(*v.InvocationResult.StandardErrorContent)
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
}
