package profiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type RemoteRegistry struct {
	URL    string `json:"url"`
	APIKey string `json:"api_key,omitempty"`
}

type RemoteClient struct {
	registry RemoteRegistry
	client   *http.Client
}

func NewRemoteClient(reg RemoteRegistry) *RemoteClient {
	return &RemoteClient{
		registry: reg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (rc *RemoteClient) Publish(profiles []Profile) error {
	body, err := json.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("marshal profiles: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", rc.registry.URL+"/profiles", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if rc.registry.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+rc.registry.APIKey)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return fmt.Errorf("publish request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (rc *RemoteClient) Subscribe() ([]Profile, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", rc.registry.URL+"/profiles", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if rc.registry.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+rc.registry.APIKey)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("subscribe request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("subscribe failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("decode profiles: %w", err)
	}
	return profiles, nil
}

type Subscription struct {
	client  *RemoteClient
	reg     *Registry
	stopCh  chan struct{}
	stopped bool
	mu      sync.Mutex
}

func (r *Registry) Subscribe(reg RemoteRegistry, interval time.Duration) *Subscription {
	sub := &Subscription{
		client: NewRemoteClient(reg),
		reg:    r,
		stopCh: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				profiles, err := sub.client.Subscribe()
				if err != nil {
					continue
				}
				for i := range profiles {
					r.Register(&profiles[i])
				}
			case <-sub.stopCh:
				return
			}
		}
	}()

	return sub
}

func (s *Subscription) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.stopped {
		close(s.stopCh)
		s.stopped = true
	}
}
