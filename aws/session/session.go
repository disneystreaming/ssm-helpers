package session

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"github.com/disneystreaming/ssm-helpers/aws/config"
)

// NewPoolSafe is used to create a pool of AWS sessions with different profile/region permutations
func NewPoolSafe(profiles []string, regions []string, logger *log.Logger) (allSessions *PoolSafe) {
	sessions := map[string]*Pool{}

	for _, region := range regions {
		if stsCredentialsSet() {
			pool := newPool("", region, logger)
			poolName := fmt.Sprintf("default-%s", region)
			sessions[poolName] = pool

			continue
		}

		for _, profile := range profiles {
			pool := newPool(profile, region, logger)
			poolName := fmt.Sprintf("%s-%s", profile, region)
			sessions[poolName] = pool
		}
	}

	return &PoolSafe{Sessions: sessions}
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

func newPool(profile string, region string, logger *log.Logger) *Pool {
	options := session.Options{
		Config:            *config.NewDefaultConfig(region),
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(options)
	if err != nil {
		logger.Fatalf("Error when trying to create session:\n%v", err)
	}

	pool := &Pool{
		Logger:      logger,
		ProfileName: profile,
		Session:     session,
	}

	return pool
}
