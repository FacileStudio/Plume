package nuage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const requestTimeout = 30 * time.Second

// StatusError is a non-success answer from Nuage.
type StatusError struct {
	Status int
	Body   string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("nuage answered %d: %s", e.Status, e.Body)
}

// Retryable reports whether trying again could change the outcome: server
// errors and rate limiting do, other client errors do not.
func (e *StatusError) Retryable() bool {
	return e.Status >= 500 || e.Status == http.StatusTooManyRequests
}

// Client talks to the Nuage file API with a restricted app token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client

	mu     sync.Mutex
	spaces map[string]int64
}

// NewClient builds a client for the Nuage instance at baseURL.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: requestTimeout},
		spaces:  make(map[string]int64),
	}
}

// EnsureSpace finds or creates the Nuage space carrying facileID and returns
// its numeric id. The answer is memoised for the life of the process.
func (c *Client) EnsureSpace(ctx context.Context, facileID, name string) (int64, error) {
	c.mu.Lock()
	id, ok := c.spaces[facileID]
	c.mu.Unlock()
	if ok {
		return id, nil
	}

	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/api/spaces/by-facile-id/"+url.PathEscape(facileID), bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return 0, statusError(resp)
	}

	var space struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&space); err != nil {
		return 0, fmt.Errorf("decode space: %w", err)
	}
	if space.ID == 0 {
		return 0, fmt.Errorf("nuage returned a space without an id")
	}

	c.mu.Lock()
	c.spaces[facileID] = space.ID
	c.mu.Unlock()
	return space.ID, nil
}

// Upload stores data in the given space under linkedTo. Nuage deduplicates on
// (space, app, linked_to), so a 200 for an existing file counts as success.
func (c *Client) Upload(ctx context.Context, spaceID int64, filename, linkedTo string, data []byte) error {
	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	part, err := form.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := form.WriteField("space_id", fmt.Sprint(spaceID)); err != nil {
		return err
	}
	if err := form.WriteField("linked_to", linkedTo); err != nil {
		return err
	}
	if err := form.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/files", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return statusError(resp)
	}
	return nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.http.Do(req)
}

func statusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return &StatusError{Status: resp.StatusCode, Body: strings.TrimSpace(string(body))}
}
