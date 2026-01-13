package config

import (
	"os"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Version   string `toml:"ver"`
	Company   string `toml:"company"`
	AppName   string `toml:"app_name"`
	MachineID string `toml:"machineID"`

	Language struct {
		English struct {
			NovelStarted string `toml:"novelStarted"`
		} `toml:"english"`
	} `toml:"LANGUGAE"`

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
			Path     string `toml:"path"`
		} `toml:"PterodactylAutoSuspend"`
	} `toml:"PLUGINS"`
}

func LoadConfig(path string) (*Config, error) {
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

func GetConfigPath() string {
	if path := os.Getenv("SENTINEL_CONFIG"); path != "" {
		return path
	}
	return "/etc/sentinel/config.toml"
}