package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/disneystreaming/ssm-helpers/cmd/builder"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	"github.com/disneystreaming/ssm-helpers/cmd/logutil"
)

var version = "devel"
var commit = "notpassed"
var rootCmd = newRootCmd()
var log = &logrus.Logger{
	Formatter: &logrus.TextFormatter{
		DisableTimestamp: true,
		PadLevelText:     true,
		ForceColors:      true,
	},
	Hooks: make(logrus.LevelHooks),
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssm",
		Short: "ssm is a user-friendly wrapper for the AWS CLI SSM commands",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			setLogLevel(cmd, args)
			logutil.SetLogSplitOutput(log)
		},
		Version: fmt.Sprintf("%s\ngit commit hash %s", version, commit),
	}

	cmdutil.AddProfileFlag(cmd)
	cmdutil.AddRegionFlag(cmd)
	cmdutil.AddInstanceFlag(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmdutil.AddVerboseFlag(cmd)
	cmdutil.AddAllProfilesFlag(cmd)
	cmdutil.AddFilterFlag(cmd)

	cmdgroup := &builder.SubCommandGroup{
		Commands: []*cobra.Command{
			newCommandSSMRun(),
			newCommandSSMSession(),
		},
	}

	cmdgroup.AddGroup(cmd)
	return cmd
}

func setLogLevel(cmd *cobra.Command, args []string) {
	log.Level = logutil.IntToLogLevel(cmdutil.GetFlagInt(cmd, "verbose"))
}

// Execute provides an entrypoint into the commands from main()
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
