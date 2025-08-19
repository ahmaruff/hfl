package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL    = "https://api.notion.com/v1"
	APIVersion = "2022-06-28"
)

type Client struct {
	token  string
	client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) makeRequest(method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	var jsonData []byte
	var err error

	if body != nil {
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		// fmt.Printf("📤 Request Body to %s %s:\n%s\n", method, url, string(jsonData))

		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, BaseURL+url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", APIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		errorBody, _ := io.ReadAll(resp.Body)

		// fmt.Printf("❌ API Error Response:\n%s\n", string(errorBody))
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(errorBody))
	}

	return resp, nil
}

func (c *Client) QueryDatabase(databaseID string, filter interface{}) (*QueryResponse, error) {
	body := map[string]interface{}{}
	if filter != nil {
		body["filter"] = filter
	}

	resp, err := c.makeRequest("POST", "/databases/"+databaseID+"/query", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetDatabase(databaseID string) (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/databases/"+databaseID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode database: %w", err)
	}
	return result, nil
}

func (c *Client) UpdateDatabase(databaseID string, properties map[string]interface{}) error {
	body := map[string]interface{}{
		"properties": properties,
	}
	resp, err := c.makeRequest("PATCH", "/databases/"+databaseID, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) CreatePage(databaseID string, properties Properties, children []Block) (*Page, error) {
	body := map[string]interface{}{
		"parent": map[string]string{
			"database_id": databaseID,
		},
		"properties": properties,
	}

	if len(children) > 0 {
		body["children"] = children
	}

	resp, err := c.makeRequest("POST", "/pages", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdatePage(pageID string, properties Properties) (*Page, error) {
	body := map[string]interface{}{
		"properties": properties,
	}

	resp, err := c.makeRequest("PATCH", "/pages/"+pageID, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetBlockChildren(blockID string) (*BlockListResponse, error) {
	resp, err := c.makeRequest("GET", "/blocks/"+blockID+"/children", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BlockListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateBlockChildren(blockID string, children []Block) error {
	existing, err := c.GetBlockChildren(blockID)
	if err != nil {
		return err
	}

	for _, block := range existing.Results {
		c.makeRequest("DELETE", "/blocks/"+block.ID, nil)
	}

	if len(children) > 0 {
		body := map[string]interface{}{
			"children": children,
		}

		_, err := c.makeRequest("PATCH", "/blocks/"+blockID+"/children", body)
		return err
	}

	return nil
}
