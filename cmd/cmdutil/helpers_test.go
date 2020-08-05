package cmdutil

import (
	"reflect"
	"testing"

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

func TestUsageError(t *testing.T) {
	assert := assert.New(t)

	t.Run("check formatted output", func(t *testing.T) {
		cmd := NewTestCmd()
		err := UsageError(cmd, "This is an error with a value: %v", "foo")
		assert.Error(err)
		assert.Equal("This is an error with a value: foo\nSee 'test -h' for help and examples", err.Error())
	})
}

func TestValidateArgs(t *testing.T) {
	assert := assert.New(t)

	t.Run("no extra args", func(t *testing.T) {
		cmd := NewTestCmd()
		err := ValidateArgs(cmd, []string{})
		assert.NoError(err)
	})

	t.Run("extra args present", func(t *testing.T) {
		cmd := NewTestCmd()
		err := ValidateArgs(cmd, []string{"foo"})
		assert.Error(err)
	})
}

func TestValidateMaxConcurrency(t *testing.T) {
	assert := assert.New(t)

	t.Run("No input is false", func(t *testing.T) {
		actual := ValidateMaxConcurrency("")
		assert.False(actual)
	})

	t.Run("Input len greater than 7 char is false", func(t *testing.T) {
		actual1 := ValidateMaxConcurrency("12345678")
		assert.False(actual1)

		actual2 := ValidateMaxConcurrency("1234567%")
		assert.False(actual2)
	})

	t.Run("Incorrect format is false", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10%10")
		assert.False(actual)
	})

	t.Run("Number is true", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10")
		assert.True(actual)
	})

	t.Run("Percentage is true", func(t *testing.T) {
		actual := ValidateMaxConcurrency("10%")
		assert.True(actual)
	})
}

func TestValidateMaxErrors(t *testing.T) {
	assert := assert.New(t)

	t.Run("No input is false", func(t *testing.T) {
		actual := ValidateMaxErrors("")
		assert.False(actual)
	})

	t.Run("Input len greater than 7 char is false", func(t *testing.T) {
		actual1 := ValidateMaxErrors("12345678")
		assert.False(actual1)

		actual2 := ValidateMaxErrors("1234567%")
		assert.False(actual2)
	})

	t.Run("Incorrect format is false", func(t *testing.T) {
		actual := ValidateMaxErrors("10%10")
		assert.False(actual)
	})

	t.Run("Number is true", func(t *testing.T) {
		actual := ValidateMaxErrors("10")
		assert.True(actual)
	})

	t.Run("Percentage is true", func(t *testing.T) {
		actual := ValidateMaxErrors("10%")
		assert.True(actual)
	})
}

func TestAddProfileFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddProfileFlag(cmd)
		assert.NotNil(cmd.Flag("profile"))
	})
}

func TestAddRegionFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddRegionFlag(cmd)
		assert.NotNil(cmd.Flag("region"))
	})
}

func TestAddFilterFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddFilterFlag(cmd)
		assert.NotNil(cmd.Flag("filter"))
	})
}

func TestAddDryRunFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddDryRunFlag(cmd)
		assert.NotNil(cmd.Flag("dry-run"))
	})
}

func TestAddVerboseFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddVerboseFlag(cmd)
		assert.NotNil(cmd.Flag("verbose"))
	})
}

func TestAddLimitFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddLimitFlag(cmd, 10, "Set a limit for the number of instances returned")

		assert.NotNil(cmd.Flag("limit"))
		assert.Equal("10", cmd.Flag("limit").DefValue)
		assert.Equal("Set a limit for the number of instances returned", cmd.Flag("limit").Usage)
	})
}

func TestAddInstanceFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddInstanceFlag(cmd)
		assert.NotNil(cmd.Flag("instance"))
	})
}

func TestAddAllProfilesFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddAllProfilesFlag(cmd)
		assert.NotNil(cmd.Flag("all-profiles"))
	})
}

func TestAddCommandFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddCommandFlag(cmd)
		assert.NotNil(cmd.Flag("command"))
	})
}

func TestAddFileFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddFileFlag(cmd, "Use this to define the path to a file")
		assert.NotNil(cmd.Flag("file"))
		assert.Equal("Use this to define the path to a file", cmd.Flag("file").Usage)
	})
}

func TestAddTagFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddTagFlag(cmd)
		assert.NotNil(cmd.Flag("tag"))
	})
}

func TestAddSessionNameFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddSessionNameFlag(cmd, "mySession")
		assert.NotNil(cmd.Flag("session-name"))
		assert.Equal("mySession", cmd.Flag("session-name").DefValue)
	})
}

func TestAddMaxConcurrencyFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddMaxConcurrencyFlag(cmd, "50", "Use this flag to set maximum concurrency")
		assert.NotNil(cmd.Flag("max-concurrency"))
		assert.Equal("50", cmd.Flag("max-concurrency").DefValue)
		assert.Equal("Use this flag to set maximum concurrency", cmd.Flag("max-concurrency").Usage)
	})
}

func TestAddMaxErrorsFlag(t *testing.T) {
	assert := assert.New(t)

	t.Run("verify flag exists", func(t *testing.T) {
		cmd := NewTestCmd()
		AddMaxErrorsFlag(cmd, "0", "Use this flag to set maximum errors")
		assert.NotNil(cmd.Flag("max-errors"))
		assert.Equal("0", cmd.Flag("max-errors").DefValue)
		assert.Equal("Use this flag to set maximum errors", cmd.Flag("max-errors").Usage)
	})
}

func TestGetCommandFlagStringSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag present, empty", func(t *testing.T) {
		cmd := NewTestCmd()
		AddCommandFlag(cmd)

		cs, err := GetCommandFlagStringSlice(cmd)
		assert.NoError(err)
		assert.Empty(cs)
	})

	t.Run("flag present, multiple commands", func(t *testing.T) {
		cmd := NewTestCmd()
		AddCommandFlag(cmd)
		cmd.SetArgs([]string{"-c", "uname -a; hostname"})
		cmd.Execute()

		cs, err := GetCommandFlagStringSlice(cmd)
		assert.NoError(err)
		assert.Len(cs, 2)
	})

	t.Run("flag not defined", func(t *testing.T) {
		cmd := NewTestCmd()

		cs, err := GetCommandFlagStringSlice(cmd)
		assert.Error(err)
		assert.Empty(cs)
	})
}

func TestGetFlagStringSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag present, empty", func(t *testing.T) {
		cmd := NewTestCmd()
		AddProfileFlag(cmd)

		cs, err := GetFlagStringSlice(cmd, "profile")
		assert.NoError(err)
		assert.Empty(cs)
	})

	t.Run("flag present, multiple commands", func(t *testing.T) {
		cmd := NewTestCmd()
		AddProfileFlag(cmd)
		cmd.SetArgs([]string{"-p", "profile1,profile2"})
		cmd.Execute()

		cs, err := GetFlagStringSlice(cmd, "profile")
		assert.NoError(err)
		assert.Len(cs, 2)
	})

	t.Run("flag not defined", func(t *testing.T) {
		cmd := NewTestCmd()

		cs, err := GetFlagStringSlice(cmd, "profile")
		assert.Error(err)
		assert.Empty(cs)
	})
}

func TestGetFlagString(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag present, empty", func(t *testing.T) {
		cmd := NewTestCmd()
		AddCommandFlag(cmd)

		cs, err := GetFlagString(cmd, "command")
		assert.NoError(err)
		assert.Empty(cs)
	})

	t.Run("flag present, multiple commands", func(t *testing.T) {
		cmd := NewTestCmd()
		AddCommandFlag(cmd)
		cmd.SetArgs([]string{"-c", "myString"})
		cmd.Execute()

		cs, err := GetFlagString(cmd, "command")
		assert.NoError(err)
		assert.Equal("myString", cs)
	})

	t.Run("flag not defined", func(t *testing.T) {
		cmd := NewTestCmd()

		cs, err := GetFlagString(cmd, "command")
		assert.Error(err)
		assert.Empty(cs)
	})
}

func TestGetFlagBool(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag present, empty", func(t *testing.T) {
		cmd := NewTestCmd()
		AddDryRunFlag(cmd)

		flag, err := GetFlagBool(cmd, "dry-run")
		assert.NoError(err)
		assert.Empty(flag)
	})

	t.Run("flag present, set to true", func(t *testing.T) {
		cmd := NewTestCmd()
		AddDryRunFlag(cmd)
		cmd.SetArgs([]string{"--dry-run"})
		cmd.Execute()

		flag, err := GetFlagBool(cmd, "dry-run")
		assert.NoError(err)
		assert.True(flag)
	})

	t.Run("flag not defined", func(t *testing.T) {
		cmd := NewTestCmd()

		flag, err := GetFlagBool(cmd, "dry-run")
		assert.Error(err)
		assert.Empty(flag)
	})
}

func TestGetFlagInt(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag present, default value", func(t *testing.T) {
		cmd := NewTestCmd()
		AddVerboseFlag(cmd)
		cmd.Execute()

		flag, err := GetFlagInt(cmd.Root(), "verbose")
		assert.NoError(err)
		assert.Equal(2, flag)
	})

	t.Run("flag present, set to 3", func(t *testing.T) {
		cmd := NewTestCmd()
		AddVerboseFlag(cmd)
		cmd.SetArgs([]string{"-v", "3"})
		cmd.Execute()

		flag, err := GetFlagInt(cmd, "verbose")
		assert.NoError(err)
		assert.Equal(3, flag)
	})

	t.Run("flag not defined", func(t *testing.T) {
		cmd := NewTestCmd()
		cmd.Execute()

		flag, err := GetFlagInt(cmd, "verbose")
		assert.Error(err)
		assert.Empty(flag)
	})
}

func TestGetMapFromStringSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("flag undefined", func(t *testing.T) {
		cmd := NewTestCmd()

		valMap, err := GetMapFromStringSlice(cmd, "fakeFlag")
		assert.Nil(valMap)
		assert.Error(err)
	})

	t.Run("malformed params when trying to squash CSV", func(t *testing.T) {
		cmd := NewTestCmd()
		AddFilterFlag(cmd)
		cmd.SetArgs([]string{"-f", "foobar"})
		cmd.Execute()

		valMap, err := GetMapFromStringSlice(cmd, "filter")
		assert.Nil(valMap)
		assert.Error(err)
	})

	t.Run("multiple values for single key", func(t *testing.T) {
		cmd := NewTestCmd()
		AddFilterFlag(cmd)
		cmd.SetArgs([]string{"-f", "foo=bar", "-f", "baz"})
		cmd.Execute()

		valMap, err := GetMapFromStringSlice(cmd, "filter")
		assert.NoError(err)

		testMap := map[string]string{
			"foo": "bar,baz",
		}

		assert.True(reflect.DeepEqual(valMap, testMap))
	})

	t.Run("multiple kv pairs", func(t *testing.T) {
		cmd := NewTestCmd()
		AddFilterFlag(cmd)
		cmd.SetArgs([]string{"-f", "foo=bar,baz=bat"})
		cmd.Execute()

		valMap, err := GetMapFromStringSlice(cmd, "filter")
		assert.NoError(err)

		testMap := map[string]string{
			"foo": "bar",
			"baz": "bat",
		}

		assert.True(reflect.DeepEqual(valMap, testMap))
	})
}

func Test_squashParamsSlice(t *testing.T) {
	assert := assert.New(t)

	t.Run("empty param slice", func(t *testing.T) {
		cmd := NewTestCmd()
		ps, err := squashParamsSlice([]string{}, cmd)

		assert.Empty(ps)
		assert.NoError(err)
	})

	t.Run("merge multiple params", func(t *testing.T) {
		cmd := NewTestCmd()
		ps, err := squashParamsSlice([]string{"env=dev", "prod"}, cmd)

		assert.Len(ps, 1)
		assert.Equal([]string{"env=dev,prod"}, ps)
		assert.NoError(err)
	})

	t.Run("malformed param", func(t *testing.T) {
		cmd := NewTestCmd()
		ps, err := squashParamsSlice([]string{"foobar"}, cmd)

		assert.Empty(ps)
		assert.Error(err)
	})
}

func Test_readAsSSV(t *testing.T) {
	assert := assert.New(t)

	t.Run("empty input", func(t *testing.T) {
		ssv := readAsSSV("")
		assert.Empty(ssv)
	})

	t.Run("splittable input", func(t *testing.T) {
		ssv := readAsSSV("foo;bar;baz")
		assert.ElementsMatch(ssv, []string{"foo", "bar", "baz"})
	})
}
