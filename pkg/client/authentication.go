package client

import (
	"context"
	"fmt"
	"net/http"
)

type basicAuthRoundTripper struct {
	username string
	password string
	rt       http.RoundTripper
}

// RoundTrip is required to implement RoundTripper interface.
// We check if an authorization token is already set, if not we set it.
// We return the initial RoundTripper to chain it in the whole process.
func (b *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		resp, err := b.rt.RoundTrip(req)
		if err != nil {
			return nil, fmt.Errorf("inner round trip error: %w", err)
		}

		return resp, nil
	}

	req = req.Clone(context.TODO())
	req.SetBasicAuth(b.username, b.password)

	resp, err := b.rt.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("inner round trip error: %w", err)
	}

	return resp, nil
}

// NewBasicAuthRoundTripper returns a basicAuthRoundTripper with username and password.
func NewBasicAuthRoundTripper(username, password string, rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = &http.Transport{}
	}

	return &basicAuthRoundTripper{
		username: username,
		password: password,
		rt:       rt,
	}
}
