package httpx

import (
	"net/http"
	"time"
)

// NewDefaultClient returns a new http client
// configured with sane timeout because globals are bad
func NewDefaultClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}
