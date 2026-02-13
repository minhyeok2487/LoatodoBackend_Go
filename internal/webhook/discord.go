package webhook

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// DiscordWebhook sends messages to Discord webhook URLs.
type DiscordWebhook struct {
	webhookURL string
	noticeURL  string
	memberURL  string
	client     *http.Client
}

// NewDiscordWebhook creates a new DiscordWebhook with the given webhook URLs.
func NewDiscordWebhook(webhookURL, noticeURL, memberURL string) *DiscordWebhook {
	return &DiscordWebhook{
		webhookURL: webhookURL,
		noticeURL:  noticeURL,
		memberURL:  memberURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// SendError sends an error message to the main webhook channel (fire-and-forget).
func (d *DiscordWebhook) SendError(message string) {
	go d.send(d.webhookURL, message)
}

// SendNotice sends a notice message to the notice webhook channel (fire-and-forget).
func (d *DiscordWebhook) SendNotice(message string) {
	go d.send(d.noticeURL, message)
}

// SendMemberNotification sends a member-related notification (fire-and-forget).
func (d *DiscordWebhook) SendMemberNotification(message string) {
	go d.send(d.memberURL, message)
}

// send posts a message to a Discord webhook URL.
func (d *DiscordWebhook) send(url, content string) {
	if url == "" {
		return
	}

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal discord payload", "error", err)
		return
	}

	resp, err := d.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to send discord webhook", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("discord webhook returned non-2xx status", "status", resp.StatusCode)
	}
}
