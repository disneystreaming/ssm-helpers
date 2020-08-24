package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"

	"github.com/disneystreaming/gomux"

	"github.com/disneystreaming/ssm-helpers/aws/session"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	ssmx "github.com/disneystreaming/ssm-helpers/ssm"
	"github.com/disneystreaming/ssm-helpers/ssm/instance"
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

	addBaseFlags(cmd)
	addSessionFlags(cmd)

	return cmd
}

func startSessionCommand(cmd *cobra.Command, args []string) {
	var err error
	var instanceList, profileList, regionList, tagList []string

	// Get all of our CLI flag values
	if err = cmdutil.ValidateArgs(cmd, args); err != nil {
		log.Fatal(err)
	}

	if instanceList, err = cmdutil.GetFlagStringSlice(cmd, "instance"); err != nil {
		log.Fatal(err)
	}

	var filterList map[string]string
	if filterList, err = cmdutil.GetMapFromStringSlice(cmd, "filter"); err != nil {
		log.Fatal(err)
	}

	if err = validateSessionFlags(cmd, instanceList, filterList); err != nil {
		log.Fatal(err)
	}

	if profileList, err = getProfileList(cmd); err != nil {
		log.Fatal(err)
	}
	if regionList, err = getRegionList(cmd); err != nil {
		log.Fatal(err)
	}
	if tagList, err = cmdutil.GetFlagStringSlice(cmd, "tag"); err != nil {
		log.Fatal(err)
	}

	var dryRunFlag bool
	if dryRunFlag, err = cmdutil.GetFlagBool(cmd, "dry-run"); err != nil {
		log.Fatal(err)
	}

	var sessionName string
	if sessionName, err = cmdutil.GetFlagString(cmd, "session-name"); err != nil {
		log.Fatal(err)
	}

	var limitFlag int
	if limitFlag, err = cmdutil.GetFlagInt(cmd, "limit"); err != nil {
		log.Fatal(err)
	}

	// Get the number of cores available for parallelization
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create our instance input object (filters, instances)
	diiInput := ssmx.CreateSSMDescribeInstanceInput(filterList, instanceList)

	// Create threadsafe pool of instance info to use for selection
	instancePool := instance.InstanceInfoSafe{
		AllInstances: make(map[string]instance.InstanceInfo),
	}

	var totalInstances int32
	var wg sync.WaitGroup

	// Set up our AWS session for each permutation of profile + region and iterate over them
	sessionPool := session.NewPool(profileList, regionList, log)
	for _, sess := range sessionPool.Sessions {
		wg.Add(1)
		go func(sess *session.Session, instancePool *instance.InstanceInfoSafe) {
			defer wg.Done()

			client := ssm.New(sess.Session)
			sessionInstances, err := instance.GetSessionInstances(client, diiInput)
			if err != nil {
				log.Tracef("AWS Session Parameters: %s, %s", *sess.Session.Config.Region, sess.ProfileName)
				log.Fatal(err)
			}

			atomic.AddInt32(&totalInstances, int32(len(sessionInstances)))
			ssmx.CheckInstanceReadiness(sess, client, sessionInstances, limitFlag, instancePool)
		}(sess, &instancePool)
	}

	wg.Wait()

	log.Infof("Retrieved %d usable instances.", len(instancePool.AllInstances))

	// No functional results, exit now
	if len(instancePool.AllInstances) == 0 || dryRunFlag {
		return
	}

	// Single instance specified or found, starting session in current terminal (non-multiplexed)
	if len(instancePool.AllInstances) == 1 {
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
					log.Fatalf("Failed to start session for instance %s\n%s", v.InstanceID, err)
				}
			}
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
		log.Info("To force nested tmux sessions, unset $TMUX.")
		log.Infof("Attach to the session with `tmux attach -t %s`", sessionName)
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

func startSelectionPrompt(instances *instance.InstanceInfoSafe, totalInstances int32, tags ssmx.ListSlice) (selectedInstances []instance.InstanceInfo, err error) {
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
