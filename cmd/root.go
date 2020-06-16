package cmd

import (
	"fmt"
	"os"

	"github.com/disneystreaming/ssm-helpers/cmd/builder"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssm",
		Short: "ssm is a user-friendly wrapper for the AWS CLI SSM commands",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	cmdutil.AddProfileFlag(cmd)
	cmdutil.AddRegionFlag(cmd)
	cmdutil.AddInstanceFlag(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmdutil.AddVerboseFlag(cmd)
	cmdutil.AddVersionFlag(cmd)
	cmdutil.AddAllProfilesFlag(cmd)
	cmdutil.AddFilterFlag(cmd)

	cmdgroup := &builder.SubCommandGroup{
		Commands: []*cobra.Command{
			newCommandSSMRun(),
		},
	}

	cmdgroup.AddGroup(cmd)
	return cmd
}

var rootCmd = newRootCmd()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
