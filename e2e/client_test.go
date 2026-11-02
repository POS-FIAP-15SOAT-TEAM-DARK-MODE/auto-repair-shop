package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// baseURL is the only address the suite talks to: the gateway. Every service is reached through it.
func baseURL() string {
	if v := os.Getenv("E2E_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:8080"
}

// reply is the last HTTP answer received by the suite.
type reply struct {
	status int
	body   []byte
}

// decode unmarshals the JSON body of the reply into v.
func (r *reply) decode(v any) error {
	if err := json.Unmarshal(r.body, v); err != nil {
		// The body is not echoed: login replies carry tokens and lambda requests carry CPFs.
		return fmt.Errorf("decode body (status %d, %d bytes): %w", r.status, len(r.body), err)
	}
	return nil
}

// httpClient sends one request through the gateway. role selects the bearer token ("" = no Authorization header).
type httpClient struct {
	base string
	http *http.Client
}

func newHTTPClient() *httpClient {
	return &httpClient{base: baseURL(), http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *httpClient) send(method, path, token string, body any) (*reply, error) {
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, c.base+path, payload)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: read body: %w", method, path, err)
	}
	return &reply{status: res.StatusCode, body: raw}, nil
}

// poll runs check every interval until it returns nil or the timeout expires. It returns the last error.
// Used where the saga updates state asynchronously (events on SQS).
func poll(timeout, interval time.Duration, check func() error) error {
	deadline := time.Now().Add(timeout)
	for {
		err := check()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("condition not reached within %s: %w", timeout, err)
		}
		time.Sleep(interval)
	}
}
