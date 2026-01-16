package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// APIClient handles communication with the remote API
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAllURLs fetches all URLs from the API
func (c *APIClient) GetAllURLs() ([]CustomURL, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/urls")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	return parseURLsFromResponse(apiResp.Data)
}

// CreateURL creates a new short URL via the API
func (c *APIClient) CreateURL(originalURL string) (*CustomURL, error) {
	reqBody, _ := json.Marshal(CreateURLRequest{URL: originalURL})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/urls", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("%s", apiResp.Error)
	}

	return parseURLFromResponse(apiResp.Data)
}

// DeleteURL deletes a URL via the API
func (c *APIClient) DeleteURL(id uint) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/urls/%d", c.BaseURL, id), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("%s", apiResp.Error)
	}

	return nil
}

// SearchURLs searches URLs via the API
func (c *APIClient) SearchURLs(query string) ([]CustomURL, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/search?q=" + url.QueryEscape(query))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API error: %s", apiResp.Error)
	}

	return parseURLsFromResponse(apiResp.Data)
}

// CheckHealth checks if the API is reachable
func (c *APIClient) CheckHealth() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("cannot connect to API at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API health check failed: %s", string(body))
	}

	return nil
}

// parseURLsFromResponse converts API response data to []CustomURL
func parseURLsFromResponse(data interface{}) ([]CustomURL, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var urlResponses []URLResponse
	if err := json.Unmarshal(jsonData, &urlResponses); err != nil {
		return nil, err
	}

	urls := make([]CustomURL, len(urlResponses))
	for i, ur := range urlResponses {
		urls[i] = urlResponseToCustomURL(ur)
	}

	return urls, nil
}

// parseURLFromResponse converts API response data to *CustomURL
func parseURLFromResponse(data interface{}) (*CustomURL, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var ur URLResponse
	if err := json.Unmarshal(jsonData, &ur); err != nil {
		return nil, err
	}

	url := urlResponseToCustomURL(ur)
	return &url, nil
}

// urlResponseToCustomURL converts URLResponse to CustomURL
func urlResponseToCustomURL(ur URLResponse) CustomURL {
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", ur.CreatedAt)
	return CustomURL{
		ID:        ur.ID,
		OldURL:    ur.OriginalURL,
		ShortURL:  ur.ShortURL,
		Visits:    ur.Visits,
		CreatedAt: createdAt,
	}
}
