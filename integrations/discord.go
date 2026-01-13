package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"anti-abuse-go/config"
	"anti-abuse-go/logger"
)

type DiscordWebhook struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Color       int            `json:"color"`
	Fields      []DiscordField `json:"fields"`
	Timestamp   string         `json:"timestamp"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

func SendDiscordWebhook(cfg *config.Config, machineID, title, description string, fields []DiscordField, aiAnalysis string) error {
	if !cfg.Integration.Discord.Enabled {
		return nil
	}

	// Prepend machine ID to fields so it's immediately visible
	machineField := DiscordField{
		Name:   "Machine ID",
		Value:  machineID,
		Inline: true,
	}
	fields = append([]DiscordField{machineField}, fields...)

	embed := DiscordEmbed{
		Title:       title,
		Description: description,
		Color:       0xff0000, // Red for alerts
		Fields:      fields,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if aiAnalysis != "" && cfg.Integration.Discord.TruncateText {
		// Truncate if needed
		if len(aiAnalysis) > 2000 {
			aiAnalysis = aiAnalysis[:1997] + "..."
		}
		embed.Description += "\n\n**AI Analysis:** " + aiAnalysis
	}

	webhook := DiscordWebhook{
		Embeds: []DiscordEmbed{embed},
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(cfg.Integration.Discord.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("discord webhook failed with status: %d", resp.StatusCode)
	}

	logger.Log.Info("Discord webhook sent")
	return nil
}
