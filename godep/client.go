// Package godep provides Go methods and structures for talking to individual DEP API endpoints.
package godep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	depclient "github.com/micromdm/nanodep/client"
)

const (
	mediaType = "application/json;charset=UTF8"
	UserAgent = "nanodep-g-o-dep/0"
)

// HTTPError encapsulates an HTTP response error from the DEP requests.
// The API returns error information in the request body.
type HTTPError struct {
	Body       []byte
	Status     string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("DEP HTTP error: %s: %s", e.Status, string(e.Body))
}

// NewHTTPError creates and returns a new HTTPError from r. Note this reads
// all of r.Body and the caller is responsible for closing it.
func NewHTTPError(r *http.Response) error {
	body, readErr := io.ReadAll(r.Body)
	err := &HTTPError{
		Body:       body,
		Status:     r.Status,
		StatusCode: r.StatusCode,
	}
	if readErr != nil {
		return fmt.Errorf("reading body of DEP HTTP error: %v: %w", err, readErr)
	}
	return err
}

// httpErrorContains checks if err is an HTTPError and contains body and a
// matching status code. With the depsim DEP simulator the body strings are
// returned with surrounding quotes. i.e. `"INVALID_CURSOR"` vs. just
// `INVALID_CURSOR` so we search the body data for the string vs. matching.
func httpErrorContains(err error, status int, s string) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == status && bytes.Contains(httpErr.Body, []byte(s)) {
		return true
	}
	return false
}

// ClientStorage provides the required data needed to connect to the Apple DEP APIs.
type ClientStorage interface {
	depclient.AuthTokensRetriever
	depclient.ConfigRetriever
}

// Client represents an Apple DEP API client identified by a single DEP name.
type Client struct {
	store  ClientStorage
	client *http.Client // for DEP API authentication and session management
	ua     string       // HTTP User-Agent
}

// Options change the configuration of the godep Client.
type Option func(*Client)

// WithUserAgent sets the the HTTP User-Agent string to be used on each request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.ua = ua
	}
}

// WithClient configures the HTTP client to be used.
// The provided client is copied and modified by wrapping its
// transport in a new NanoDEP transport (which transparently handles
// authentication and session management). If not set then
// http.DefaultClient is used.
func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// NewClient creates new Client and reads authentication and config data from store.
func NewClient(store ClientStorage, opts ...Option) *Client {
	c := &Client{
		store:  store,
		client: http.DefaultClient,
		ua:     UserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	t := depclient.NewTransport(c.client.Transport, c.client, store, nil)
	c.client = depclient.NewClient(c.client, t)
	return c
}

// do executes the HTTP request using the client's HTTP client which
// should be using the NanoDEP transport (which handles authentication).
// This frees us to only be concerned about the actual DEP API request.
// We encode in to JSON and decode any returned body as JSON to out.
func (c *Client) do(ctx context.Context, name, method, path string, in interface{}, out interface{}) error {
	var body io.Reader
	if in != nil {
		bodyBytes, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := depclient.NewRequestWithContext(ctx, name, c.store, method, path, body)
	if err != nil {
		return err
	}
	if c.ua != "" {
		req.Header.Set("User-Agent", c.ua)
	}
	if body != nil {
		req.Header.Set("Content-Type", mediaType)
	}
	if out != nil {
		req.Header.Set("Accept", mediaType)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unhandled auth error: %w", depclient.NewAuthError(resp))
	} else if resp.StatusCode != http.StatusOK {
		return NewHTTPError(resp)
	}

	if out != nil {
		err := json.NewDecoder(resp.Body).Decode(out)
		if err != nil {
			return err
		}
	}

	return nil
}
