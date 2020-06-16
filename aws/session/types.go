package session

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
)

// Pool is a type that holds an instance of an AWS session as well as the profile name used to initialize it
type Pool struct {
	Session     *session.Session
	ProfileName string
}

// PoolSafe is a type that allows for goroutine-safe access to a slice of SessionPool objects via mutex locks
type PoolSafe struct {
	sync.Mutex
	Sessions map[string]*Pool
}
