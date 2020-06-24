package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultHTTPClient(t *testing.T) {
	c := NewDefaultClient()
	assert.Equal(t, 10*time.Second, c.Timeout)
}
