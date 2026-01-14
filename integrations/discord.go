package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
	Author      *DiscordAuthor `json:"author,omitempty"`
	Thumbnail   *DiscordImage  `json:"thumbnail,omitempty"`
}

type DiscordAuthor struct {
	Name string `json:"name"`
}

type DiscordImage struct {
	URL string `json:"url"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

func SendDiscordWebhook(cfg *config.Config, machineID, filePath string, fields []DiscordField, aiAnalysis string) error {
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
		Title:       fmt.Sprintf("Sentinel Detection Alert - %s", machineID),
		Description: aiAnalysis,
		Color:       65280, // Green for alerts
		Fields:      fields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Author: &DiscordAuthor{
			Name: filePath,
		},
	}

	// Truncate description if needed
	if len(embed.Description) > 4096 {
		embed.Description = embed.Description[:4093] + "..."
	}

	webhook := DiscordWebhook{
		Embeds: []DiscordEmbed{embed},
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	// Check if file exists and is under 10MB for attachment
	var body io.Reader
	var contentType string
	if stat, err := os.Stat(filePath); err == nil && stat.Size() < 10*1024*1024 {
		// Create multipart form
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add payload_json
		payloadField, _ := writer.CreateFormField("payload_json")
		payloadField.Write(data)

		// Add file
		fileField, _ := writer.CreateFormFile("file", filepath.Base(filePath))
		file, _ := os.Open(filePath)
		defer file.Close()
		io.Copy(fileField, file)

		writer.Close()
		body = &buf
		contentType = writer.FormDataContentType()
	} else {
		body = bytes.NewBuffer(data)
		contentType = "application/json"
	}

	resp, err := http.Post(cfg.Integration.Discord.WebhookURL, contentType, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return fmt.Errorf("discord webhook failed with status: %d", resp.StatusCode)
	}

	logger.Log.Info("Discord webhook sent")
	return nil
}
