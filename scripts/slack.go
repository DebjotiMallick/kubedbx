package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	slackSuccessColor = "#36a64f" // green
	slackFailureColor = "#dc3545" // red
	slackWarningColor = "#ffc107" // yellow
)

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color string `json:"color"`
	Text  string `json:"text"`
}

func sendSlackNotification(message SlackMessage) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL not set")
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack notification failed with status: %d", resp.StatusCode)
	}

	return nil
}
