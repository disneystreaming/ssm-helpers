package session

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/disneystreaming/ssm-helpers/aws/config"
)

// NewPoolSafe is used to create a pool of AWS sessions with different profile/region permutations
func NewPoolSafe(profiles []string, regions []string) (allSessions *PoolSafe) {

	wg := sync.WaitGroup{}
	sp := &PoolSafe{
		Sessions: make(map[string]*Pool),
	}

	if regions != nil {
		wg.Add(len(profiles) * len(regions))
		for _, p := range profiles {
			for _, r := range regions {
				// Wait until we have the session for each permutation of profiles and regions
				go func(p string, r string) {
					defer wg.Done()

					newSession := newSession(p, r)
					sp.Lock()

					session := Pool{
						Session:     newSession,
						ProfileName: p,
					}
					sp.Sessions[fmt.Sprintf("%s-%s", p, r)] = &session
					//sp.Sessions = append(sp.Sessions, &session)
					defer sp.Unlock()
				}(p, r)
			}
		}
	} else {
		wg.Add(len(profiles))
		for _, p := range profiles {
			// Wait until we have the session for each profile
			go func(p string) {
				defer wg.Done()

				newSession := newSession(p, "")
				sp.Lock()

				session := Pool{
					Session:     newSession,
					ProfileName: p,
				}
				sp.Sessions[fmt.Sprintf("%s", p)] = &session

				//sp.Sessions = append(sp.Sessions, &session)
				defer sp.Unlock()
			}(p)
		}
	}

	// Wait until all sessions have been initialized
	wg.Wait()

	return sp
}

// createSession uses a given profile and region to call NewSessionWithOptions() to initialize an instance of the AWS client with the given settings.
// If the region is nil, it defaults to the default region in the ~/.aws/config file or the AWS_REGION environment variable.
func newSession(profile string, region string) (newSession *session.Session) {
	// Create AWS session from shared config
	// This will import the AWS_PROFILE envvar from your console, if set
	if region != "" {
		newSession = session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Config:            *config.NewDefaultConfigWithRegion(region),
					Profile:           profile,
					SharedConfigState: session.SharedConfigEnable,
				}))
	} else {
		newSession = session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Config:            *config.NewDefaultConfig(),
					Profile:           profile,
					SharedConfigState: session.SharedConfigEnable,
				}))
	}

	return newSession
}
