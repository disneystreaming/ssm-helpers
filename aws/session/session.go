package session

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"github.com/disneystreaming/ssm-helpers/aws/config"
)

// NewPool is used to create a pool of AWS sessions with different profile/region permutations
func NewPool(profiles []string, regions []string, logger *log.Logger) *Pool {
	sessions := map[string]*Session{}

	for _, region := range regions {
		if stsCredentialsSet() {
			session := newSession("", region, logger)
			name := fmt.Sprintf("default-%s", region)
			sessions[name] = session

			continue
		}

		for _, profile := range profiles {
			session := newSession(profile, region, logger)
			name := fmt.Sprintf("%s-%s", profile, region)
			sessions[name] = session
		}
	}

	return &Pool{Sessions: sessions}
}

func stsCredentialsSet() bool {
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if sessionToken == "" || accessKeyID == "" || secretAccessKey == "" {
		return false
	}

	return true
}

func newSession(profile string, region string, logger *log.Logger) *Session {
	options := session.Options{
		Config:            *config.NewDefaultConfig(region),
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(options)
	if err != nil {
		logger.Fatalf("Error when trying to create session:\n%v", err)
	}

	return &Session{
		Logger:      logger,
		ProfileName: profile,
		Session:     session,
	}
}
