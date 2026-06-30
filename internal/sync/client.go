package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	ServerURL string
	Token     string
	http      *http.Client
}

func NewClient(serverURL, token string) *Client {
	return &Client{
		ServerURL: serverURL,
		Token:     token,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

type SyncRecord struct {
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type PushRequest struct {
	Changes []SyncRecord `json:"changes"`
}

type PullRequest struct {
	Since string `json:"since"`
}

type PullResponse struct {
	Changes    []SyncRecord `json:"changes"`
	ServerTime string       `json:"server_time"`
}

func (c *Client) Push(changes []SyncRecord) error {
	body, _ := json.Marshal(PushRequest{Changes: changes})
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("push failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) Pull(since string) (*PullResponse, error) {
	body, _ := json.Marshal(PullRequest{Since: since})
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/v1/sync/pull", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pull failed: %s", resp.Status)
	}

	var result PullResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}
