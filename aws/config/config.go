package config

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"

	"github.com/disneystreaming/ssm-helpers/util/httpx"
)

// NewDefaultConfig will return default AWS Configs that are meant to be
// merged in. Luckily both the deprecated session.New() and session.NewSession()
// take `cfgs ...*aws.Config` then merge them in.
//
// This means we can change our sessions to be `session.New(<whatver>, session.NewDefaultConfig())
// If we need to override it we can swap order (last config's value wins)
func NewDefaultConfig(region string) *aws.Config {
	return &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(region),
		HTTPClient:                    httpx.NewDefaultClient(),
		Retryer: &client.DefaultRetryer{
			NumMaxRetries:    3,
			MaxThrottleDelay: 1500 * time.Millisecond,
		},
	}
}
