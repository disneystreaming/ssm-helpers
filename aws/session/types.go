package session

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

// Session is an initialized session configuration for a target profile and region
type Session struct {
	Logger      *logrus.Logger
	Session     *session.Session
	ProfileName string
}

// Pool contains a set of session configurations for the target profiles and regions
type Pool struct {
	Sessions map[string]*Session
}
