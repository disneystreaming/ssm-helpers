package session

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"github.com/disneystreaming/ssm-helpers/aws/config"
)

// NewPoolSafe is used to create a pool of AWS sessions with different profile/region permutations
func NewPoolSafe(profiles []string, regions []string, logger *log.Logger) (allSessions *PoolSafe) {

	wg := sync.WaitGroup{}
	sp := &PoolSafe{
		Sessions: make(map[string]*Pool),
	}

	if len(regions) != 0 {
		for _, p := range profiles {
			for _, r := range regions {
				wg.Add(1)
				// Wait until we have the session for each permutation of profiles and regions
				go func(p string, r string) {
					defer wg.Done()

					s, err := newSession(p, r)
					if err != nil {
						logger.Fatalf("Error when trying to create session:\n%v", err)
					}

					if err = validateSessionCreds(s); err != nil {
						logger.Fatal(err)
					}

					session := Pool{
						Logger:      logger,
						ProfileName: p,
						Session:     s,
					}
					sp.Sessions[fmt.Sprintf("%s-%s", p, r)] = &session
				}(p, r)
			}
		}
	} else {
		for _, p := range profiles {
			wg.Add(1)
			// Wait until we have the session for each profile
			go func(p string) {
				defer wg.Done()

				s, err := newSession(p, "")
				if err != nil {
					logger.Fatalf("Error when trying to create session:\n%v", err)
				}

				if err = validateSessionCreds(s); err != nil {
					logger.Fatal(err)
				}

				session := Pool{
					Logger:      logger,
					ProfileName: p,
					Session:     s,
				}
				sp.Sessions[fmt.Sprintf("%s", p)] = &session
			}(p)
		}
	}
	// Wait until all sessions have been initialized
	wg.Wait()

	return sp
}

func validateSessionCreds(session *session.Session) (err error) {
	creds := session.Config.Credentials
	if _, err := creds.Get(); err != nil {
		return fmt.Errorf("Error when validating credentials:\n%v", err)
	}

	return nil
}

// newSession uses a given profile and region to call NewSessionWithOptions() to initialize an instance of the AWS client with the given settings.
// If the region is nil, it defaults to the default region in the ~/.aws/config file or the AWS_REGION environment variable.
func newSession(profile string, region string) (newSession *session.Session, err error) {
	// Create AWS session from shared config
	// This will import the AWS_PROFILE envvar from your console, if set
	return session.NewSessionWithOptions(
		session.Options{
			Config:            *config.NewDefaultConfig(region),
			Profile:           profile,
			SharedConfigState: session.SharedConfigEnable,
		})
}
