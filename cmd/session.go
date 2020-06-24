package cmd

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
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"

	"github.com/disneystreaming/gomux"
	awsx "github.com/disneystreaming/ssm-helpers/aws"
	"github.com/disneystreaming/ssm-helpers/aws/session"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	ssmx "github.com/disneystreaming/ssm-helpers/ssm"
	"github.com/disneystreaming/ssm-helpers/ssm/instance"
	"github.com/disneystreaming/ssm-helpers/util"
)

func newCommandSSMSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "open a terminal session to an instance using SSM",
		Long:  "foo bar baz",
		Run: func(cmd *cobra.Command, args []string) {
			startSessionCommand(cmd, args)
		},
	}

	cmdutil.AddLimitFlag(cmd, 10, "Set a limit for the number of instance results returned per profile/region combination.")
	cmdutil.AddTagFlag(cmd)
	cmdutil.AddSessionNameFlag(cmd, "ssm-session")
	return cmd
}

func startSessionCommand(cmd *cobra.Command, args []string) {
	cmdutil.ValidateArgs(cmd, args)

	dryRunFlag := cmdutil.GetFlagBool(cmd.Parent(), "dry-run")
	profileList := cmdutil.GetFlagStringSlice(cmd.Parent(), "profile")
	regionList := cmdutil.GetFlagStringSlice(cmd.Parent(), "region")
	filterList := cmdutil.GetFlagStringSlice(cmd.Parent(), "filter")
	tagList := cmdutil.GetFlagStringSlice(cmd, "tag")
	limitFlag := cmdutil.GetFlagInt(cmd, "limit")
	instanceList := cmdutil.GetFlagStringSlice(cmd, "instance")
	sessionName := cmdutil.GetFlagString(cmd, "session-name")

	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	var wg sync.WaitGroup

	// Master list for later
	instancePool := instance.InstanceInfoSafe{
		AllInstances: make(map[string]instance.InstanceInfo),
	}
	var totalInstances int
	// Iterate through our AWS sessions
	wg.Add(len(sessionPool.Sessions))
	for _, sess := range sessionPool.Sessions {

		go func(sess *session.Pool, instancePool *instance.InstanceInfoSafe) {
			defer wg.Done()

			instanceChan := make(chan []*ssm.InstanceInformation)
			errChan := make(chan error)
			svc := ssm.New(sess.Session)

			go ssmx.GetInstanceList(svc, filterMaps, instanceList, false, instanceChan, errChan)
			instanceList, err := <-instanceChan, <-errChan

			if err != nil {
				log.Debugf("AWS Session Parameters: %s, %s", *sess.Session.Config.Region, sess.ProfileName)
				log.Error(err)
			}

			totalInstances += len(instanceList)
			ssmx.CheckInstanceReadiness(sess, svc, instanceList, instancePool, limitFlag)
		}(sess, &instancePool)
	}

	wg.Wait()

	log.Infof("Found %d usable instances.", len(instancePool.AllInstances))

	// No functional results, exit now
	if len(instancePool.AllInstances) == 0 {
		return
	}

	// If -i flag is set, don't prompt for instance selection
	if !dryRunFlag {
		// Single instance specified or found, starting session in current terminal (non-multiplexed)
		if len(instanceList) == 1 {
			for _, v := range instancePool.AllInstances {
				if err := startSSMSession(v.Profile, v.Region, v.InstanceID); err != nil {
					log.Errorf("Failed to start ssm-session for instance %s\n%s", v.InstanceID, err)
				}
			}
			return
		}

		// Multiple instances specified or found, check to see if we're in a tmux session to avoid nesting
		if len(instanceList) > 1 && len(instancePool.AllInstances) > 1 {
			var instances []instance.InstanceInfo
			for _, v := range instancePool.AllInstances {
				instances = append(instances, v)
			}

			if err := configTmuxSession(sessionName, instances); err != nil {
				log.Fatal(err)
			}
		} else {
			// If -i was not specified, go to a selection prompt before starting sessions
			selectedInstances, err := startSelectionPrompt(&instancePool, totalInstances, tagList)
			if err != nil {
				if err == terminal.InterruptErr {
					log.Info("Instance selection interrupted.")
					os.Exit(0)
				}
				log.Errorf("Error during instance selection\n%s", err)
				os.Exit(1)
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

func startSelectionPrompt(instances *instance.InstanceInfoSafe, totalInstances int, tags ssmx.ListSlice) (selectedInstances []instance.InstanceInfo, err error) {
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
