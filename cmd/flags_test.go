package cmd

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/disneystreaming/ssm-helpers/cmd/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func NewTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			return
		},
	}
	return cmd
}

func Test_getProfileList(t *testing.T) {
	assert := assert.New(t)
	cmd := NewTestCmd()

	t.Run("profile flag undefined", func(t *testing.T) {
		cmd.Execute()

		profileList, err := getProfileList(cmd)
		assert.Nil(profileList)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("allProfiles flag undefined", func(t *testing.T) {
		cmdutil.AddProfileFlag(cmd)
		cmd.Execute()

		profileList, err := getProfileList(cmd)
		assert.Nil(profileList)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("try to use --profile and --all-profiles", func(t *testing.T) {
		cmdutil.AddProfileFlag(cmd)
		cmdutil.AddAllProfilesFlag(cmd)

		cmd.SetArgs([]string{"-p", "profile1", "--all-profiles"})
		cmd.Execute()

		profileList, err := getProfileList(cmd)
		assert.Nil(profileList)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("profile from envvar", func(t *testing.T) {
		err := os.Setenv("AWS_PROFILE", "testProfile")
		assert.NoError(err, "error when trying to set AWS_PROFILE envvar\n%v", err)

		cmdutil.AddProfileFlag(cmd)
		cmdutil.AddAllProfilesFlag(cmd)
		cmd.SetArgs([]string{})
		cmd.Execute()

		profileList, err := getProfileList(cmd)

		assert.Len(profileList, 1)
		assert.Contains(profileList, "testProfile")
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("profile fallback to default", func(t *testing.T) {
		err := os.Unsetenv("AWS_PROFILE")
		assert.NoError(err, "error when trying to unset AWS_PROFILE envvar\n%v", err)

		cmdutil.AddProfileFlag(cmd)
		cmdutil.AddAllProfilesFlag(cmd)
		cmd.SetArgs([]string{})
		cmd.Execute()

		profileList, err := getProfileList(cmd)

		assert.Len(profileList, 1)
		assert.Contains(profileList, "default")
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("single profile", func(t *testing.T) {
		cmdutil.AddProfileFlag(cmd)
		cmdutil.AddAllProfilesFlag(cmd)

		cmd.SetArgs([]string{"-p", "myProfile"})
		cmd.Execute()

		profileList, err := getProfileList(cmd)
		assert.Len(profileList, 1)
		assert.NoError(err)

		cmd.ResetFlags()

	})

	t.Run("multiple profiles", func(t *testing.T) {
		cmdutil.AddProfileFlag(cmd)
		cmdutil.AddAllProfilesFlag(cmd)

		cmd.SetArgs([]string{"-p", "account1,account2,account3"})
		cmd.Execute()

		profileList, err := getProfileList(cmd)
		assert.Len(profileList, 3)
		assert.NoError(err)

		cmd.ResetFlags()

	})
}

func Test_getRegionList(t *testing.T) {
	assert := assert.New(t)
	cmd := NewTestCmd()

	t.Run("region flag undefined", func(t *testing.T) {
		cmd.Execute()

		regionList, err := getRegionList(cmd)
		assert.Len(regionList, 0)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("region from envvar", func(t *testing.T) {
		cmdutil.AddRegionFlag(cmd)
		cmd.Execute()

		os.Setenv("AWS_REGION", "us-east-1")
		regionList, err := getRegionList(cmd)

		assert.Len(regionList, 1)
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("single region", func(t *testing.T) {
		cmdutil.AddRegionFlag(cmd)

		cmd.SetArgs([]string{"-r", "us-east-1"})
		cmd.Execute()

		regionList, err := getRegionList(cmd)
		assert.Len(regionList, 1)
		assert.NoError(err)

		cmd.ResetFlags()

	})

	t.Run("multiple regions", func(t *testing.T) {
		cmdutil.AddRegionFlag(cmd)

		cmd.SetArgs([]string{"-r", "us-east-1,us-west-2,eu-central-1"})
		cmd.Execute()

		regionList, err := getRegionList(cmd)
		assert.Len(regionList, 3)
		assert.NoError(err)

		cmd.ResetFlags()

	})
}

func Test_getFilterList(t *testing.T) {
	assert := assert.New(t)
	cmd := NewTestCmd()

	t.Run("filter flag undefined", func(t *testing.T) {
		cmd.Execute()

		filterList, err := getFilterList(cmd)
		assert.Len(filterList, 0)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("single filter", func(t *testing.T) {
		cmdutil.AddFilterFlag(cmd)
		cmd.SetArgs([]string{"-f", "foo=bar"})
		cmd.Execute()

		filterList, err := getFilterList(cmd)
		assert.Len(filterList, 1)
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("multiple filters", func(t *testing.T) {
		cmdutil.AddFilterFlag(cmd)
		cmd.SetArgs([]string{"-f", "foo=bar,baz=bat"})
		cmd.Execute()

		filterList, err := getFilterList(cmd)
		assert.Len(filterList, 2)
		assert.NoError(err)

		cmd.ResetFlags()
	})
}

func Test_getCommandList(t *testing.T) {
	assert := assert.New(t)
	cmd := NewTestCmd()

	t.Run("command flag undefined", func(t *testing.T) {
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 0)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("file flag undefined", func(t *testing.T) {
		cmdutil.AddCommandFlag(cmd)
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 0)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("single command", func(t *testing.T) {
		addRunFlags(cmd)
		cmd.SetArgs([]string{"-c", "hostname"})
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 1)
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("multiple commands", func(t *testing.T) {
		addRunFlags(cmd)
		cmd.SetArgs([]string{"-c", "uname -a; hostname"})
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 2)
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("using --file only", func(t *testing.T) {
		addRunFlags(cmd)
		cmd.SetArgs([]string{"--file", "../testing/test_commands.sh"})
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 5)
		assert.NoError(err)

		cmd.ResetFlags()
	})

	t.Run("invalid --file", func(t *testing.T) {
		addRunFlags(cmd)
		cmd.SetArgs([]string{"--file", "../testing/does_not_exist.sh"})
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Nil(commandList)
		assert.Error(err)

		cmd.ResetFlags()
	})

	t.Run("using --file and --command", func(t *testing.T) {
		addRunFlags(cmd)
		cmd.SetArgs([]string{"--file", "../testing/test_commands.sh", "-c", "whoami"})
		cmd.Execute()

		commandList, err := getCommandList(cmd)
		assert.Len(commandList, 6)
		assert.NoError(err)

		cmd.ResetFlags()
	})
}

func Test_validateRunFlags(t *testing.T) {
	assert := assert.New(t)

	cmd := NewTestCmd()
	cmdutil.AddCommandFlag(cmd)
	cmdutil.AddFilterFlag(cmd)
	cmdutil.AddInstanceFlag(cmd)
	cmd.Execute()

	instanceList := make([]string, 51)

	t.Run("try to use --filter and --instance flags", func(t *testing.T) {
		targetList := make([]*ssm.Target, 2)
		err := validateRunFlags(cmd, instanceList, []string{"hostname"}, targetList)
		assert.Error(err)
	})

	t.Run("no instances or filters specified", func(t *testing.T) {
		err := validateRunFlags(cmd, nil, []string{"hostname"}, nil)
		assert.Error(err)
	})

	t.Run(">50 specified instances", func(t *testing.T) {
		err := validateRunFlags(cmd, instanceList, []string{"hostname"}, nil)
		assert.Error(err)
	})

	t.Run("no command specified", func(t *testing.T) {
		err := validateRunFlags(cmd, []string{"myInstance"}, nil, nil)
		assert.Error(err)
	})

	t.Run("valid flag combination", func(t *testing.T) {
		err := validateRunFlags(cmd, []string{"myInstance"}, []string{"hostname"}, nil)
		assert.NoError(err)
	})
}

func Test_setLogLevel(t *testing.T) {
	assert := assert.New(t)

	cmd := NewTestCmd()

	testLogger := logrus.New()

	t.Run("verbose flag undefined", func(t *testing.T) {
		cmd.Execute()
		err := setLogLevel(cmd, testLogger)
		assert.Error(err)
	})

	t.Run("check that correct level is set", func(t *testing.T) {
		cmdutil.AddVerboseFlag(cmd)
		cmd.SetArgs([]string{"-v", "0"})
		cmd.Execute()

		err := setLogLevel(cmd, testLogger)
		assert.Equal(logrus.FatalLevel, testLogger.GetLevel())
		assert.NoError(err)
		cmd.ResetFlags()
	})

}
