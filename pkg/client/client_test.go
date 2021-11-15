package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/flatcar-linux/fleetlock/pkg/client"
)

var fleetlockHeaders = http.Header{
	"Fleet-Lock-Protocol": []string{"true"},
}

type httpClient struct {
	do func(req *http.Request) (*http.Response, error)
	r  *http.Request
}

func (m *httpClient) Do(req *http.Request) (*http.Response, error) {
	m.r = req

	return m.do(req)
}

func (m *httpClient) RoundTrip(req *http.Request) (*http.Response, error) {
	m.r = req

	return m.do(req)
}

func TestBadURL(t *testing.T) {
	t.Parallel()

	_, err := client.New(&client.Config{URL: "this is not an URL", ID: "1234"})
	if err == nil {
		t.Fatalf("should get an error")
	}

	expectedError := "parsing URL: parse \"this is not an URL\": invalid URI for request"
	if err != nil && err.Error() != expectedError {
		t.Fatalf("should have %s for the error, got: %s", expectedError, err.Error())
	}
}

//nolint:funlen // Just many test cases.
func TestClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, test := range []struct {
		statusCode int
		expErr     error
		body       []byte
		doErr      error
		cfg        *client.Config
		expCfg     *client.Config
	}{
		{
			statusCode: 200,
			cfg: &client.Config{
				ID:  "1234",
				URL: "http://1.2.3.4",
			},
			expCfg: &client.Config{
				Group: "default",
			},
		},
		{
			statusCode: 500,
			expErr:     errors.New("fleetlock error: this is an error (error_kind)"),
			body:       []byte(`{"kind":"error_kind","value":"this is an error"}`),
			cfg: &client.Config{
				ID:  "1234",
				URL: "http://1.2.3.4",
			},
			expCfg: &client.Config{
				Group: "default",
			},
		},
		{
			statusCode: 500,
			expErr:     errors.New("unmarshalling error: invalid character '\"' after object key:value pair"),
			body:       []byte(`{"kind":"error_kind"  "value":"this is an error"}`),
			cfg: &client.Config{
				ID:    "1234",
				URL:   "http://1.2.3.4",
				Group: "lokomotive",
			},
			expCfg: &client.Config{
				Group: "lokomotive",
			},
		},
		{
			statusCode: 100,
			expErr:     errors.New("unexpected status code: 100"),
			cfg: &client.Config{
				ID:  "1234",
				URL: "http://1.2.3.4",
			},
			expCfg: &client.Config{
				Group: "default",
			},
		},
		{
			expErr: errors.New("doing the request: connection refused"),
			doErr:  errors.New("connection refused"),
			cfg: &client.Config{
				ID:  "1234",
				URL: "http://1.2.3.4",
			},
			expCfg: &client.Config{
				Group: "default",
			},
		},
	} {
		test := test

		newClient := func(cfg *client.Config, statusCode int, body []byte, doErr error) (*httpClient, *client.Client) {
			h := &httpClient{
				do: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(bytes.NewReader(body)),
					}, doErr
				},
			}

			cfg.HTTP = h

			c, err := client.New(cfg)
			if err != nil {
				t.Fatalf("Unexpected error creating client: %v", err)
			}

			return h, c
		}

		getPayload := func(h *httpClient) *client.Payload {
			b, err := h.r.GetBody()
			if err != nil {
				t.Fatalf("unable to get body from request: %v", err)
			}

			payload, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatalf("unable to read body: %v", err)
			}

			var p client.Payload
			if err := json.Unmarshal(payload, &p); err != nil {
				t.Fatalf("unable to unmarshal payload: %v", err)
			}

			return &p
		}

		t.Run(fmt.Sprintf("UnlockIfHeld_%d", test.statusCode), func(t *testing.T) {
			t.Parallel()

			h, c := newClient(test.cfg, test.statusCode, test.body, test.doErr)

			err := c.UnlockIfHeld(ctx)
			if err != nil && err.Error() != test.expErr.Error() {
				t.Fatalf("should have %v for err, got: %v", test.expErr, err)
			}

			expURL := "http://1.2.3.4/v1/steady-state"

			if h.r.URL.String() != expURL {
				t.Fatalf("should have %s for URL, got: %s", expURL, h.r.URL.String())
			}

			if !reflect.DeepEqual(h.r.Header, fleetlockHeaders) {
				t.Fatalf("should have %v for headers, got: %s", fleetlockHeaders, h.r.Header)
			}

			payload := getPayload(h)

			if payload.ClientParams.Group != test.expCfg.Group {
				t.Fatalf("payload's group should be : %s, got: %s", test.expCfg.Group, payload.ClientParams.Group)
			}
		})

		t.Run(fmt.Sprintf("RecursiveLock_%d", test.statusCode), func(t *testing.T) {
			t.Parallel()

			h, c := newClient(test.cfg, test.statusCode, test.body, test.doErr)

			err := c.RecursiveLock(ctx)
			if err != nil && err.Error() != test.expErr.Error() {
				t.Fatalf("should have %v for err, got: %v", test.expErr, err)
			}

			expURL := "http://1.2.3.4/v1/pre-reboot"

			if h.r.URL.String() != expURL {
				t.Fatalf("should have %s for URL, got: %s", expURL, h.r.URL.String())
			}

			if !reflect.DeepEqual(h.r.Header, fleetlockHeaders) {
				t.Fatalf("should have %v for headers, got: %s", fleetlockHeaders, h.r.Header)
			}

			payload := getPayload(h)

			if payload.ClientParams.Group != test.expCfg.Group {
				t.Fatalf("payload's group should be : %s, got: %s", test.expCfg.Group, payload.ClientParams.Group)
			}
		})
	}
}

func Test_Client_use_given_context_for_requests(t *testing.T) {
	t.Parallel()

	key := struct{}{}
	value := "bar"

	h := &httpClient{
		do: func(req *http.Request) (*http.Response, error) {
			if req.Context().Value(key) == nil {
				t.Fatalf("Expected request to use given context")
			}

			return &http.Response{
				StatusCode: 200,
			}, nil
		},
	}

	c, err := client.New(&client.Config{
		URL:  "http://1.2.3.4",
		ID:   "1234",
		HTTP: h,
	})
	if err != nil {
		t.Fatalf("Unexpected error creating client: %v", err)
	}

	ctx := context.WithValue(context.Background(), key, value)

	if err := c.RecursiveLock(ctx); err != nil {
		t.Fatalf("Unexpected error while doing recursive lock: %v", err)
	}

	if err := c.UnlockIfHeld(ctx); err != nil {
		t.Fatalf("Unexpected error while unlocking: %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	t.Parallel()

	var (
		username = "flatcar"
		password = "p4ssw0rd"
	)

	ctx := context.Background()

	tr := &httpClient{
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
			}, nil
		},
	}

	h := http.Client{
		Transport: client.NewBasicAuthRoundTripper(username, password, tr),
	}

	c, err := client.New(&client.Config{ID: "1234", HTTP: &h, URL: "http://1.2.3.4"})
	if err != nil {
		t.Fatalf("Unexpected error creating client: %v", err)
	}

	err = c.RecursiveLock(ctx)
	if err != nil {
		t.Fatalf("should have nil for err, got: %v", err)
	}

	u, p, ok := tr.r.BasicAuth()
	if u != username || p != password || !ok {
		t.Fatalf("basic auth creds do not match")
	}
}

func TestRequiredParameters(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		cfg *client.Config
		err error
	}{
		{
			cfg: &client.Config{
				URL: "http://1.2.3.4",
			},
			err: errors.New("ID is required"),
		},
		{
			cfg: &client.Config{
				ID: "1234",
			},
			err: errors.New("URL is required"),
		},
	} {
		_, err := client.New(test.cfg)
		if err == nil {
			t.Fatal("error should not be nil")
		}

		if err.Error() != test.err.Error() {
			t.Fatalf("error should be: %v, got: %v", test.err, err)
		}
	}
}
