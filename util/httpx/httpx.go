package httpx

import (
	"net"
	"net/http"
	"time"
)

// NewDefaultClient returns a new http client
// configured with sane timeout because globals are bad
func NewDefaultClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).Dial,
		},
	}
}
