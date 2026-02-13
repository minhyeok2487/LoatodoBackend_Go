package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LostarkClient is a base HTTP client for the Lostark Open API.
type LostarkClient struct {
	httpClient *http.Client
}

// NewLostarkClient creates a new LostarkClient with a 30-second timeout.
func NewLostarkClient() *LostarkClient {
	return &LostarkClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Get sends a GET request to the given Lostark API URL with the specified API key.
func (c *LostarkClient) Get(url, apiKey string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

// Post sends a POST request with a JSON body to the given Lostark API URL.
func (c *LostarkClient) Post(url string, body interface{}, apiKey string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp)
}

// handleResponse reads the response body and returns an error for non-200 status codes.
func (c *LostarkClient) handleResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized:
		return nil, &LostarkAPIError{
			StatusCode: resp.StatusCode,
			Message:    "올바르지 않은 apiKey 입니다. (크롬 자동 번역을 확인해주세요.)",
		}
	case http.StatusTooManyRequests:
		return nil, &LostarkAPIError{
			StatusCode: resp.StatusCode,
			Message:    "사용한도 (1분에 100개)를 초과했습니다.",
		}
	case http.StatusServiceUnavailable:
		return nil, &LostarkAPIError{
			StatusCode: resp.StatusCode,
			Message:    "로스트아크 서버가 점검중 입니다.",
		}
	default:
		return nil, &LostarkAPIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)),
		}
	}
}

// FindEvents calls the Lostark events endpoint to validate an API key.
// GET https://developer-lostark.game.onstove.com/news/events
func (c *LostarkClient) FindEvents(apiKey string) error {
	_, err := c.Get("https://developer-lostark.game.onstove.com/news/events", apiKey)
	return err
}

// LostarkAPIError represents an error response from the Lostark API.
type LostarkAPIError struct {
	StatusCode int
	Message    string
}

func (e *LostarkAPIError) Error() string {
	return e.Message
}
