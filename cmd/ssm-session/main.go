package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"github.com/disneystreaming/go-ssmhelpers/aws"
	"github.com/disneystreaming/go-ssmhelpers/aws/session"
	ssmhelpers "github.com/disneystreaming/go-ssmhelpers/ssm"
	"github.com/disneystreaming/go-ssmhelpers/ssm/instance"
	"github.com/disneystreaming/go-ssmhelpers/util"
	"github.com/disneystreaming/gomux"
)

var myInstances ssmhelpers.CommaSlice
var myFilters ssmhelpers.CommaSlice
var myProfiles ssmhelpers.CommaSlice
var myRegions ssmhelpers.CommaSlice
var myTags ssmhelpers.ListSlice

var allProfilesFlag bool
var verboseFlag int
var dryRunFlag bool
var sessionName string
var limitFlag int
var totalInstances int
var versionFlag bool
var version = "devel"
var commit = "notpassed"

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

	// Flag to indicate that the user wants to perform a dry run
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Retrieve the list of profiles, regions, and instances on which sessions would be started.")

	// Flag to indicate that the user wants to execute a command against all of their configured profiles
	flag.BoolVar(&allProfilesFlag, "all-profiles", false, "[USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.")

	// Flag to enable increasingly-verbose output
	flag.IntVar(&verboseFlag, "log-level", 0, "Sets verbosity of output:\n0 = quiet, 1 = terse, 2 = warn, 3 = debug")

	// Flag for instance selection
	flag.VarP(&myInstances, "instances", "i", "Specify what instance IDs you want to target.\nMultiple allowed, delimited by commas (e.g. --instances i-12345,i-23456)")

	// Flags for filters
	flag.VarP(&myFilters, "filter", "f", "Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).\nMultiple allowed, delimited by commas (e.g. env=dev,foo=bar)")

	flag.VarP(&myTags, "tag", "t", "Adds the specified tag as an additional column to be displayed during the instance selection prompt.")

	// Flags for profiles/regions
	flag.VarP(&myProfiles, "profiles", "p", "Specify a specific profile to use with your API calls.\nMultiple allowed, delimited by commas (e.g. --profiles dev,myaccount)")

	flag.VarP(&myRegions, "regions", "r", "Specify a specific region to use with your API calls.\n"+
		"This option will override any profile settings in your config file.\n"+
		"Multiple allowed, delimited by commas (e.g. --regions us-east-1,us-west-2)\n\n"+
		"[NOTE] Mixing --profiles and --regions will result in your command targeting every matching instance in the selected profiles and regions.\n"+
		"e.g., \"--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1\" will target instances in each of the profile/region combinations:\n"+
		"\t\"foo@us-east-1, foo@us-west-2, foo@eu-east-1\"\n"+
		"\t\"bar@us-east-1, bar@us-west-2, bar@eu-east-1\"\n"+
		"\t\"baz@us-east-1, baz@us-west-2, baz@eu-east-1\"\n"+
		"Please be careful.")

	// Flag to allow naming of tmux session
	flag.StringVar(&sessionName, "session-name", "ssm-session", "Specify a name for the tmux session created when multiple instances are selected")

	// Flag to set a limit to the number of instances returned by the SSM/EC2 API query
	flag.IntVarP(&limitFlag, "limit", "l", 20, "Set a limit for the number of instance results returned per profile/region combination.")

	// Flag to show the version number
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")

	flag.Parse()

	if versionFlag {
		fmt.Printf("Version: %s\tGit Commit Hash: %s\n", version, commit)
		os.Exit(0)
	}

	if verboseFlag == 3 {
		log.SetLevel(log.DebugLevel)
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
		if exists {
			myRegions.Set(env)
		}
	}
	// If --all-profiles is set, we call getAWSProfiles() and iterate through the user's ~/.aws/config
	if allProfilesFlag {
		profiles, err := aws.GetAWSProfiles()
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

	var wg sync.WaitGroup

	// Master list for later
	instancePool := instance.InstanceInfoSafe{
		AllInstances: make(map[string]instance.InstanceInfo),
	}

	// Iterate through our AWS sessions
	wg.Add(len(sessionPool.Sessions))
	for _, sess := range sessionPool.Sessions {

		go func(sess *session.Pool, instancePool *instance.InstanceInfoSafe) {
			defer wg.Done()

			instanceChan := make(chan []*ssm.InstanceInformation)
			errChan := make(chan error)
			svc := ssm.New(sess.Session)

			go ssmhelpers.GetInstanceList(svc, filterMaps, myInstances, false, instanceChan, errChan)
			instanceList, err := <-instanceChan, <-errChan

			if err != nil {
				log.Debugf("AWS Session Parameters: %s, %s", *sess.Session.Config.Region, sess.ProfileName)
				log.Error(err)
			}

			totalInstances = len(instanceList)
			ssmhelpers.CheckInstanceReadiness(sess, svc, instanceList, instancePool, limitFlag)
		}(sess, &instancePool)
	}

	wg.Wait()

	if verboseFlag > 0 {
		log.Infof("Found %d usable instances.", len(instancePool.AllInstances))
	}

	// No functional results, exit now
	if len(instancePool.AllInstances) == 0 {
		return
	}

	// If -i flag is set, don't prompt for instance selection
	if !dryRunFlag {
		// Single instance specified or found, starting session in current terminal (non-multiplexed)
		if len(myInstances) == 1 {
			for _, v := range instancePool.AllInstances {
				if err := startSSMSession(v.Profile, v.Region, v.InstanceID); err != nil {
					log.Errorf("Failed to start ssm-session for instance %s\n%s", v.InstanceID, err)
				}
			}
			return
		}

		// Multiple instances specified or found, check to see if we're in a tmux session to avoid nesting
		if len(myInstances) > 1 && len(instancePool.AllInstances) > 1 {
			var instances []instance.InstanceInfo
			for _, v := range instancePool.AllInstances {
				instances = append(instances, v)
			}

			if err := configTmuxSession(sessionName, instances); err != nil {
				log.Fatal(err)
			}
		} else {
			// If -i was not specified, go to a selection prompt before starting sessions
			selectedInstances, err := startSelectionPrompt(&instancePool, totalInstances, myTags)
			if err != nil {
				log.Fatalf("Error during instance selection\n%s", err)
			}

			// If only one instance was selected, don't bother with a tmux session
			if len(selectedInstances) == 1 {
				for _, v := range selectedInstances {
					if err := startSSMSession(v.Profile, v.Region, v.InstanceID); err != nil {
						log.Errorf("Failed to start ssm-session for instance %s\n%s", v.InstanceID, err)
					}
				}
				return
			}

			if err = configTmuxSession(sessionName, selectedInstances); err != nil {
				log.Fatal(err)
			}
		}

		// Make sure we aren't going to nest tmux sessions
		currentTmuxSocket := os.Getenv("TMUX")
		if len(currentTmuxSocket) == 0 {
			if err := attachTmuxSession(sessionName); err != nil {
				log.Errorf("Could not attach to tmux session '%s'\n%s", sessionName, err)
			}
		} else {
			log.Info("To force nested Tmux sessions unset $TMUX")
			log.Infof("Attach to the session with `tmux attach -t %s`", sessionName)
		}
	}
}

func configTmuxSession(sessionName string, selectedInstances []instance.InstanceInfo) (err error) {
	// Initialize our tmux session
	tmuxSession, err := gomux.NewSession(sessionName)
	if err != nil {
		return fmt.Errorf("Failed to create tmux session\n%s", err)
	}

	// Create the window in which our ssm session panes will live
	tmuxWindow, err := tmuxSession.AddWindow("ssm")
	if err != nil {
		return fmt.Errorf("Failed to create tmux window\n%s", err)
	}

	// Configure our session-specific settings
	configList := []string{
		"set-option -t " + sessionName + " pane-border-status top",
		"set-option -t " + sessionName + " mouse on",
	}

	for _, v := range configList {
		if err = tmuxSession.Windows[0].SetConfig(v); err != nil {
			return fmt.Errorf("Failed to set tmux configuration for window\n%s", err)
		}
	}

	// Multiple instances specified or found, starting tmux session and attaching current terminal to it
	for _, v := range selectedInstances {
		// Add a window for our instance to our tmux session
		if err = addInstanceToTmuxWindow(tmuxWindow, v.Profile, v.Region, v.InstanceID); err != nil {
			return fmt.Errorf("Failed to add instance %s to tmux session\n%s", v.InstanceID, err)

		}

		// Re-tile our layout after each window to avoid the "pane too small" error
		if err = tmuxWindow.SetConfig("select-layout -t " + sessionName + " tiled"); err != nil {
			return fmt.Errorf("Failed to re-tile panes\n%s", err)
		}
	}

	// Don't kill window 0 of our current session if we're already in a session
	currentTmuxSocket := os.Getenv("TMUX")
	if len(currentTmuxSocket) == 0 {
		if err = tmuxWindow.KillPane(0); err != nil {
			return fmt.Errorf("Failed to remove empty pane at index 0\n%s", err)
		}
	}
	// Re-tile our layout one last time since we removed the empty pane
	if err = tmuxWindow.SetConfig("select-layout -t " + sessionName + " tiled"); err != nil {
		return fmt.Errorf("Failed to re-tile panes\n%s", err)
	}

	return
}

func setSessionStatus(sessionName string, status string) (err error) {
	rawCmd := exec.Command("tmux", "set-option", "-t", sessionName, "status-left", status)

	return rawCmd.Run()
}

func startSSMSession(profile string, region string, instanceID string) error {
	rawCmd := exec.Command("aws", "ssm", "start-session", "--profile", profile, "--region", region, "--target", instanceID)
	rawCmd.Stdin = os.Stdin
	rawCmd.Stdout = os.Stdout
	rawCmd.Stderr = os.Stderr

	err := rawCmd.Start()
	if err != nil {
		return err
	}

	// Set up to capture Ctrl+C
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	doneChan := make(chan struct{}, 2)

	// Run Wait() in its own chan so we don't block
	go func() {
		err = rawCmd.Wait()
		doneChan <- struct{}{}
	}()
	// Here we block until command is done
	for {
		select {
		case s := <-sigChan:
			// user typed Ctrl-C, most likley meant for ssm-session pass through
			rawCmd.Process.Signal(s)
		case <-doneChan:
			// command is done
			return err
		}
	}
	return err
}

func attachTmuxSession(sessionName string) (err error) {
	// If we don't redirect these, our console will detach when ssm-session finishes executing.
	rawCmd := exec.Command("tmux", "attach", "-t", sessionName)
	rawCmd.Stdin = os.Stdin
	rawCmd.Stdout = os.Stdout
	rawCmd.Stderr = os.Stderr

	return rawCmd.Run()
}

func addInstanceToTmuxWindow(tmuxWindow *gomux.Window, profile string, region string, instanceID string) (err error) {
	tPane, err := tmuxWindow.Pane(0).Split()
	if err != nil {
		return err
	}

	if err = tPane.SetName(instanceID); err != nil {
		return err
	}

	return tPane.Exec(fmt.Sprintf("aws ssm start-session --profile %s --region %s --target %s", profile, region, instanceID))
}

func startSelectionPrompt(instances *instance.InstanceInfoSafe, totalInstances int, tags ssmhelpers.ListSlice) (selectedInstances []instance.InstanceInfo, err error) {
	instanceIDList := []string{}
	promptList := instances.FormatStringSlice([]string(tags)...)
	fmt.Println("      ", promptList[0])

	prompt := &survey.MultiSelect{
		Message: fmt.Sprintf("Showing %d/%d instances. Make a Selection:", len(instances.AllInstances), totalInstances),
		Options: promptList[1 : len(promptList)-1],
	}

	if err := survey.AskOne(prompt, &instanceIDList, survey.WithPageSize(25)); err != nil {
		return nil, err
	}

	if len(instanceIDList) == 0 {
		return nil, fmt.Errorf("No instances selected")
	}

	// This is clunky, but currently necessary in order to finish creating the tmux sessions
	for _, v := range instanceIDList {
		id := strings.Split(v, " ")[0]
		if id != "" {
			selectedInstances = append(selectedInstances, instances.AllInstances[id])
		}
	}
	return selectedInstances, nil
}
