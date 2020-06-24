package ec2

// InstanceTags is a simple struct to hold an instanceID and a map of tag data
type InstanceTags struct {
	InstanceID string
	Tags       map[string]string
}
