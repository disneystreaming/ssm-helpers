package ec2

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocks "github.com/disneystreaming/ssm-helpers/testing"
)

func TestGetEC2InstanceTags(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockEC2Client{}

	tags, _ := GetEC2InstanceTags(mockSvc, "i-123", "i-456", "i-789")

	t.Run("instance tag slice size", func(t *testing.T) {
		assert.Lenf(tags, 3, "Incorrect number of tag slices returned, got %d, expected 3", len(tags))
		assert.Truef(tags[0].InstanceID == "i-123" && tags[1].InstanceID == "i-456" && tags[2].InstanceID == "i-789", "Incorrect instance ID returned in tag slice\n%v", tags)
	})

	t.Run("instance tag data", func(t *testing.T) {
		assert.Truef(tags[0].Tags["env"] == "dev" && tags[0].Tags["id_foo"] == "321", "Incorrect tag values returned for instance i-123, got %v, expected env:dev and id_foo:321", tags[0].Tags)
	})
}

func TestGetEC2InstanceInfo(t *testing.T) {
	assert := assert.New(t)
	mockSvc := &mocks.MockEC2Client{}

	info, _ := getEC2InstanceInfo(mockSvc, "i-123", "i-456", "i-789")

	assert.Lenf(info[0].Reservations[0].Instances, 3, "Incorrect number of instances returned, got %d, expected 3", len(info[0].Reservations[0].Instances))
}
