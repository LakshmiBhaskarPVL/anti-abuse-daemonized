package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	AppName = "Sentinel"
	Company = "Novel"
)

const DefaultConfigTemplate = `# Sentinel Configuration by Novel
ver = "1.0.0"
machineID = "node1"

[LOGS]
processStartMsg = true
flaggedNoti = true
fileModified = false
fileDeleted = false
fileMoved = false
fileCreated = false

[DETECTION]
# Multiple watchdog paths can be monitored simultaneously
watchdogPath = [
    "/var/lib/pterodactyl/volumes",
    # "/var/www/html",
    # "/root/.ssh"
]
SignaturePath = "/etc/sentinel/signatures"
watchdogIgnorePath = ["/etc/sentinel/signatures"]
watchdogIgnoreFile = ["main.go", "config.toml"]
maxFileSizeMB = 500  # Allow up to 500MB files

[INTEGRATION.AI]
enabled = true
generate_models = ["llama-3.3-70b-versatile", "llama-3.3-70b-specdec"]
generate_endpoint = "http://localhost:11434/api/generate"
use_groq = false
groq_api_token = ""
prompt = "Analyze the given code and return an abuse score (0-10) with a brief reason. Example abuses: Crypto Mining, Shell Access, Nezha Proxy (VPN/Proxy usage), Disk Filling, Tor, DDoS, Abusive Resource Usage. Response format: '**5/10** <your reason>'. No extra messages."

[INTEGRATION.DISCORD]
enabled = false
webhook_url = "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
truncate_text = true

[PLUGINS.PterodactylAutoSuspend]
hostname = "https://panel.example.com"
api_key = "ptla_"
`

type Config struct {
	Version   string `toml:"ver"`
	MachineID string `toml:"machineID"`

	Logs struct {
		ProcessStartMsg bool `toml:"processStartMsg"`
		FlaggedNoti     bool `toml:"flaggedNoti"`
		FileModified    bool `toml:"fileModified"`
		FileDeleted     bool `toml:"fileDeleted"`
		FileMoved       bool `toml:"fileMoved"`
		FileCreated     bool `toml:"fileCreated"`
	} `toml:"LOGS"`

	Detection struct {
		WatchdogPath     []string `toml:"watchdogPath"`
		SignaturePath    string   `toml:"SignaturePath"`
		WatchdogIgnorePath []string `toml:"watchdogIgnorePath"`
		WatchdogIgnoreFile []string `toml:"watchdogIgnoreFile"`
		MaxFileSizeMB    int      `toml:"maxFileSizeMB"` // Optional, default 100MB
	} `toml:"DETECTION"`

	Integration struct {
		AI struct {
			Enabled        bool     `toml:"enabled"`
			GenerateModels []string `toml:"generate_models"`
			GenerateEndpoint string `toml:"generate_endpoint"`
			UseGroq        bool     `toml:"use_groq"`
			GroqAPIKey     string   `toml:"groq_api_token"`
			Prompt         string   `toml:"prompt"`
		} `toml:"AI"`
		Discord struct {
			Enabled      bool   `toml:"enabled"`
			WebhookURL   string `toml:"webhook_url"`
			TruncateText bool   `toml:"truncate_text"`
		} `toml:"DISCORD"`
	} `toml:"INTEGRATION"`

	Plugins struct {
		PterodactylAutoSuspend struct {
			Hostname string `toml:"hostname"`
			APIKey   string `toml:"api_key"`
		} `toml:"PterodactylAutoSuspend"`
	} `toml:"PLUGINS"`
}

func LoadConfig(path string) (*Config, error) {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// If config doesn't exist, create it with defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := createDefaultConfig(path); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	// Ensure watchdogPath is a slice
	if len(config.Detection.WatchdogPath) == 0 {
		// Default if not set
		config.Detection.WatchdogPath = []string{"/var/lib/pterodactyl/volumes"}
	}

	return &config, nil
}

func createDefaultConfig(path string) error {
	return os.WriteFile(path, []byte(DefaultConfigTemplate), 0644)
}

func GetConfigPath() string {
	if path := os.Getenv("SENTINEL_CONFIG"); path != "" {
		return path
	}
	return "/etc/sentinel/config.toml"
}