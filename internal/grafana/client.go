package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	URL        string
	APIKey     string
	HTTPClient *http.Client
}

type Annotation struct {
	What string   `json:"what"`
	Tags []string `json:"tags"`
	Data string   `json:"data"`
	When int64    `json:"when"`
}

type AnnotationResponse struct {
	ID int64 `json:"id"`
}

type AnnotationPatch struct {
	TimeEnd  int64    `json:"timeEnd"`
	IsRegion bool     `json:"isRegion"`
	Tags     []string `json:"tags"`
}

func (c *Client) CreateAnnotation(ctx context.Context, what string, tags []string, data string) (int64, error) {
	payload := Annotation{What: what, Tags: tags, Data: data, When: time.Now().Unix()}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal: %w", err)
	}
	url := fmt.Sprintf("%s/api/annotations/graphite", c.URL)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return 0, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("grafana %d: %s", resp.StatusCode, string(body))
	}
	var r AnnotationResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}
	return r.ID, nil
}

func (c *Client) UpdateAnnotationToRegion(ctx context.Context, id int64, tags []string) error {
	patch := AnnotationPatch{TimeEnd: time.Now().UnixMilli(), IsRegion: true, Tags: tags}
	jsonData, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	url := fmt.Sprintf("%s/api/annotations/%d", c.URL, id)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("grafana %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
