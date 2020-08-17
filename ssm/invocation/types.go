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

// Status ...
type Status string

const (
	// CommandSuccess indicates ssm successfully ran a command on target(s)
	CommandSuccess Status = "Success"

	// CommandFailed indicates ssm ran a command on target(s) that failed
	CommandFailed Status = "Failed"

	// CommandPending indicates ssm ran a command on target(s) that is pending
	CommandPending Status = "Pending"

	// CommandInProgress The command has been sent to the instance but has not reached a terminal state.
	CommandInProgress Status = "In Progress"

	// CommandDeliveryTimedOut The command was not delivered to the instance before the delivery timeout expired. Delivery timeouts do not count against the parent command's MaxErrors limit, but they do contribute to whether the parent command status is Success or Incomplete. This is a terminal state.
	CommandDeliveryTimedOut Status = "Delivery Timed Out"

	// CommandExecutionTimedOut Command execution started on the instance, but the execution was not complete before the execution timeout expired. Execution timeouts count against the MaxErrors limit of the parent command. This is a terminal state.
	CommandExecutionTimedOut Status = "Execution Timed Out"

	// CommandCanceled The command was terminated before it was completed. This is a terminal state.
	CommandCanceled Status = "Canceled"

	// CommandUndeliverable The command can't be delivered to the instance. The instance might not exist or might not be responding. Undeliverable invocations don't count against the parent command's MaxErrors limit and don't contribute to whether the parent command status is Success or Incomplete. This is a terminal state.
	CommandUndeliverable Status = "Undeliverable"

	// CommandTerminated The parent command exceeded its MaxErrors limit and subsequent command invocations were canceled by the system. This is a terminal state.
	CommandTerminated Status = "Terminated"

	// ClientError indicates error occured prior to or when invoking an ssm api method
	ClientError Status = "ClientError"
)

// Result is used to store information about an invocation run on a particular instance
type Result struct {
	InvocationResult *ssm.GetCommandInvocationOutput
	ProfileName      string
	Region           string
	Status           Status
	Error            error
}

// ResultSafe allows for concurrent-safe access to a slice of InvocationResult info
type ResultSafe struct {
	sync.Mutex
	InvocationResults []*Result
}

// Add allows appending results safely
func (results *ResultSafe) Add(result *Result) {
	results.Lock()
	results.InvocationResults = append(results.InvocationResults, result)
	results.Unlock()
}
