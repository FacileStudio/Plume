package nuage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const requestTimeout = 30 * time.Second

// StatusError is a non-success answer from Nuage or Registre.
type StatusError struct {
	Service string
	Status  int
	Body    string
}

func (e *StatusError) Error() string {
	service := e.Service
	if service == "" {
		service = "nuage"
	}
	return fmt.Sprintf("%s answered %d: %s", service, e.Status, e.Body)
}

// Retryable reports whether trying again could change the outcome: server
// errors and rate limiting do, other client errors do not.
func (e *StatusError) Retryable() bool {
	return e.Status >= 500 || e.Status == http.StatusTooManyRequests
}

// Client talks to Nuage's app-folder API with a delegated token.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient builds a client for the Nuage instance at baseURL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: requestTimeout},
	}
}

// Put stores data in the acting user's Plume folder under key. A second Put
// with the same key replaces the object, so a retry is idempotent: 201 for a
// new key and 200 for a replaced one both count as success.
func (c *Client) Put(ctx context.Context, token, key, name string, data []byte) error {
	query := url.Values{"key": {key}, "name": {name}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/api/app/object?"+query.Encode(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/pdf")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return statusError(resp)
	}
	return nil
}

func statusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return &StatusError{Status: resp.StatusCode, Body: strings.TrimSpace(string(body))}
}
