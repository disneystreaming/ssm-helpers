package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	mocks "github.com/disneystreaming/ssm-helpers/testing"
)

func TestGetEC2InstanceTags(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockEC2Client{}

	t.Run("", func(t *testing.T) {
		tags, err := GetEC2InstanceTags(mockSvc, aws.StringSlice([]string{"i-123", "i-456", "i-789"}))

		assert.NoError(err)
		assert.Lenf(tags, 3, "Incorrect number of tag slices returned, got %d, expected 3", len(tags))
		assert.Contains(tags, []string{"i-123", "i-456", "i-789"}, "Incorrect instance ID returned in tag slice\n%v", tags)
		assert.Equal(tags["i-123"]["env"], "dev", "Incorrect tag values returned for instance i-123, got , expected env:dev and id_foo:321")

	})
}

func TestGetEC2InstanceInfo(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockEC2Client{}

	t.Run("get instance info from IDs", func(t *testing.T) {
		info, err := getEC2InstanceInfo(mockSvc,
			aws.StringSlice([]string{"i-123", "i-456", "i-789"}))

		assert.Lenf(info, 3, "Incorrect number of instances returned, got %d, expected 3", len(info))
		assert.NoError(err)
	})

}
