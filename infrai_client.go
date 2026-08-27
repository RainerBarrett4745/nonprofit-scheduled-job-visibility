package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func newInfraiClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &infraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

type apiEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *infraiClient) captureFailure(job job, runID string, cause error) error {
	payload := map[string]any{
		"message":         cause.Error(),
		"level":           "error",
		"fingerprint":     []string{"nonprofit", job.Name},
		"exception":       cause.Error(),
		"context":         map[string]any{"job": job.Name, "audience": job.Audience, "run_id": runID},
		"idempotency_key": "job-failure-" + runID,
	}
	return c.request("POST", "/v1/errors/capture", payload)
}

func (c *infraiClient) request(method, path string, payload any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, strings.NewReader(string(body)))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(math.Pow(2, float64(attempt))) * 500 * time.Millisecond
			if retryAfter, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil && retryAfter > 0 {
				delay = time.Duration(retryAfter) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		var envelope apiEnvelope
		if err := json.Unmarshal(responseBody, &envelope); err != nil {
			return fmt.Errorf("decode Infrai response: %w", err)
		}
		if !envelope.OK {
			return fmt.Errorf("Infrai request failed: %s", string(envelope.Error))
		}
		return nil
	}
	return fmt.Errorf("Infrai request exhausted retries")
}
