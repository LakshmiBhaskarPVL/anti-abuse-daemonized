package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"anti-abuse-go/config"
	"anti-abuse-go/logger"
)

type AIAnalysis struct {
	Score   int    `json:"score"`
	Reason  string `json:"reason"`
	Content string `json:"content"`
}

func AnalyzeWithAI(cfg *config.Config, content string) (*AIAnalysis, error) {
	if !cfg.Integration.AI.Enabled {
		return nil, nil
	}

	for _, model := range cfg.Integration.AI.GenerateModels {
		analysis, err := callAI(cfg, model, content)
		if err == nil {
			return analysis, nil
		}
		logger.Log.WithError(err).Warnf("AI model %s failed, trying next", model)
	}

	return nil, fmt.Errorf("all AI models failed")
}

func callAI(cfg *config.Config, model, content string) (*AIAnalysis, error) {
	prompt := fmt.Sprintf(cfg.Integration.AI.Prompt, content)

	var payload map[string]interface{}
	var url string

	if cfg.Integration.AI.UseGroq {
		url = "https://api.groq.com/openai/v1/chat/completions"
		payload = map[string]interface{}{
			"model": model,
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
			"temperature": 0.1,
			"max_tokens":  512,
		}
	} else {
		url = cfg.Integration.AI.GenerateEndpoint
		payload = map[string]interface{}{
			"model":  model,
			"prompt": prompt,
			"stream": false,
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if cfg.Integration.AI.UseGroq {
		req.Header.Set("Authorization", "Bearer "+cfg.Integration.AI.GroqAPIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(body))
	}

	return parseAIResponse(cfg, body)
}

func extractGroqContent(response map[string]interface{}) (string, error) {
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in Groq response")
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}
	msg, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no message in choice")
	}
	content, ok := msg["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content in message")
	}
	return content, nil
}

func parseAIResponse(cfg *config.Config, body []byte) (*AIAnalysis, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse AI response JSON: %w", err)
	}

	var content string
	if cfg.Integration.AI.UseGroq {
		var err error
		content, err = extractGroqContent(response)
		if err != nil {
			logger.Log.WithError(err).Warnf("Failed to extract Groq response content")
			content = fmt.Sprintf("Error in AI response: %v", err)
		}
	} else {
		if resp, ok := response["response"].(string); ok {
			content = resp
		}
	}

	if content == "" {
		return &AIAnalysis{Content: "No content in AI response"}, nil
	}

	// Parse score and reason from response like "**5/10** reason"
	parts := strings.SplitN(content, "**", 3)
	if len(parts) < 3 {
		return &AIAnalysis{Content: content}, nil
	}

	scorePart := strings.Trim(parts[1], "/10 ")
	score := 0
	if _, err := fmt.Sscanf(scorePart, "%d", &score); err != nil {
		logger.Log.WithError(err).Warnf("Failed to parse abuse score from: %s", scorePart)
		score = 0
	}

	reason := strings.TrimSpace(parts[2])

	return &AIAnalysis{
		Score:   score,
		Reason:  reason,
		Content: content,
	}, nil
}