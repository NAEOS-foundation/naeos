// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

//go:build integration && !nobroker

package broker

import (
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestRealNATSIntegrationPubSub(t *testing.T) {
	serverURL := os.Getenv("NAEOS_NATS_TEST_URL")
	if serverURL == "" {
		t.Skip("set NAEOS_NATS_TEST_URL to a running NATS server, for example nats://127.0.0.1:4222")
	}

	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse NAEOS_NATS_TEST_URL: %v", err)
	}
	if parsedURL.Scheme != "nats" || parsedURL.Hostname() == "" || parsedURL.Path != "" || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		t.Fatalf("NAEOS_NATS_TEST_URL must be a nats://host:port URL without a path, query, or fragment")
	}
	if parsedURL.User != nil && parsedURL.User.Username() != "" {
		t.Fatal("NATS integration test adapter supports a password but not a username")
	}

	port := 4222
	if parsedURL.Port() != "" {
		port, err = strconv.Atoi(parsedURL.Port())
		if err != nil || port < 1 || port > 65535 {
			t.Fatalf("invalid NATS server port %q", parsedURL.Port())
		}
	}
	password := ""
	if parsedURL.User != nil {
		password, _ = parsedURL.User.Password()
	}

	client := NewRealNATS()
	if err := client.Connect(&Config{
		Host:     parsedURL.Hostname(),
		Port:     port,
		Password: password,
		Timeout:  3 * time.Second,
	}); err != nil {
		t.Fatalf("connect to NATS server: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close NATS client: %v", err)
		}
	})

	if err := client.Ping(); err != nil {
		t.Fatalf("ping connected NATS server: %v", err)
	}

	subject := "naeos.integration." + generateID()
	received := make(chan *Message, 1)
	if err := client.Subscribe(subject, func(msg *Message) error {
		received <- msg
		return nil
	}); err != nil {
		t.Fatalf("subscribe to %q: %v", subject, err)
	}

	payload := []byte("naeos real NATS integration test")
	if err := client.Publish(subject, &Message{Payload: payload}); err != nil {
		t.Fatalf("publish to %q: %v", subject, err)
	}

	select {
	case msg := <-received:
		if msg.Channel != subject {
			t.Errorf("received channel = %q, want %q", msg.Channel, subject)
		}
		if string(msg.Payload) != string(payload) {
			t.Errorf("received payload = %q, want %q", msg.Payload, payload)
		}
		if msg.ID == "" {
			t.Error("received message has an empty ID")
		}
		if msg.Timestamp.IsZero() {
			t.Error("received message has a zero timestamp")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for NATS message on %q", subject)
	}

	if err := client.Unsubscribe(subject); err != nil {
		t.Fatalf("unsubscribe from %q: %v", subject, err)
	}
}
