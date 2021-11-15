// Package client implements FleetLock client.
package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// HTTPClient interface holds the required Post method
// to send FleetLock requests.
type HTTPClient interface {
	// Do send a `body` payload to the URL.
	Do(*http.Request) (*http.Response, error)
}

// Payload is the content to send
// to the FleetLock server.
type Payload struct {
	// ClientParams holds the parameters specific to the
	// FleetLock client.
	//
	//nolint:tagliatelle // FleetLock protocol requires exactly 'client_params' field.
	ClientParams *Params `json:"client_params"`
}

// Params is the object holding the
// ID and the group for each client.
type Params struct {
	// ID is the client identifier. (e.g node name or UUID)
	ID string `json:"id"`
	// Group is the reboot-group of the client.
	Group string `json:"group"`
}

// Client holds the params related to the host
// in order to interact with the FleetLock server URL.
type Client struct {
	baseServerURL string
	group         string
	id            string
	http          HTTPClient
}

// New builds a FleetLock client.
func New(cfg *Config) (*Client, error) {
	fleetlock := &Client{
		baseServerURL: cfg.URL,
		http:          cfg.HTTP,
		group:         cfg.Group,
		id:            cfg.ID,
	}

	if fleetlock.id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	if fleetlock.baseServerURL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	if _, err := url.ParseRequestURI(fleetlock.baseServerURL); err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	if fleetlock.group == "" {
		fleetlock.group = "default"
	}

	if fleetlock.http == nil {
		fleetlock.http = http.DefaultClient
	}

	return fleetlock, nil
}

// RecursiveLock tries to reserve (lock) a slot for rebooting.
func (c *Client) RecursiveLock(ctx context.Context) error {
	req, err := c.generateRequest(ctx, "v1/pre-reboot")
	if err != nil {
		return fmt.Errorf("generating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("doing the request: %w", err)
	}

	return handleResponse(resp)
}

// UnlockIfHeld tries to release (unlock) a slot that it was previously holding.
func (c *Client) UnlockIfHeld(ctx context.Context) error {
	req, err := c.generateRequest(ctx, "v1/steady-state")
	if err != nil {
		return fmt.Errorf("generating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("doing the request: %w", err)
	}

	return handleResponse(resp)
}

func (c *Client) generateRequest(ctx context.Context, endpoint string) (*http.Request, error) {
	payload := &Payload{
		ClientParams: &Params{
			ID:    c.id,
			Group: c.group,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling the payload: %w", err)
	}

	j := bytes.NewReader(body)

	targetURL := fmt.Sprintf("%s/%s", c.baseServerURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, j)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	headers := make(http.Header)
	headers.Add("fleet-lock-protocol", "true")
	req.Header = headers

	return req, nil
}

func handleResponse(resp *http.Response) error {
	maxHTTPErrorCode := 600

	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		return nil
	case code >= 300 && code < maxHTTPErrorCode:
		// We try to extract an eventual error.
		r := bufio.NewReader(resp.Body)

		body, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("reading body: %w", err)
		}

		//nolint:errcheck // We do it best effort and at least stdlib client never returns error here.
		resp.Body.Close()

		e := &Error{}
		if err := json.Unmarshal(body, &e); err != nil {
			return fmt.Errorf("unmarshalling error: %w", err)
		}

		return fmt.Errorf("fleetlock error: %s", e.String())
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
