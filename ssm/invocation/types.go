package invocation

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/ssm"
)

// RunShellScriptParameters is the appropriate struct for use when setting SSM document parameters
type RunShellScriptParameters map[string][]*string

// CommandOutputSafe allows for concurrent-safe access to a slice of *ssm.SendCommandOutput info
type CommandOutputSafe struct {
	sync.Mutex
	Output []*ssm.SendCommandOutput
}

// Result is used to store information about an invocation run on a particular instance
type Result struct {
	InvocationResult *ssm.GetCommandInvocationOutput
	ProfileName      string
	Region           string
	Status           string
}

// ResultSafe allows for concurrent-safe access to a slice of InvocationResult info
type ResultSafe struct {
	sync.Mutex
	InvocationResults []*Result
}
