package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"anti-abuse-go/config"
	"anti-abuse-go/logger"
)

type PterodactylAutoSuspend struct {
	cfg *config.Config
}

func init() {
	RegisterPlugin(&PterodactylAutoSuspend{})
}

func (p *PterodactylAutoSuspend) Name() string {
	return "Pterodactyl Auto Suspend"
}

func (p *PterodactylAutoSuspend) Version() string {
	return "1.0.0"
}

func (p *PterodactylAutoSuspend) OnStart(cfg *config.Config) error {
	p.cfg = cfg
	logger.Log.Info("Pterodactyl Auto Suspend plugin started")
	return nil
}

func (p *PterodactylAutoSuspend) OnDetected(path string, matches interface{}) error {
	uuid := p.extractUUID(path)
	if uuid == "" {
		return nil
	}

	serverID, err := p.getServerID(uuid)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get server ID")
		return err
	}

	return p.suspendServer(serverID)
}

func (p *PterodactylAutoSuspend) OnScan(path string, content []byte, eventType string) error {
	// No action needed
	return nil
}

var uuidRegex = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

func (p *PterodactylAutoSuspend) extractUUID(path string) string {
	// Extract UUID from path like /var/lib/pterodactyl/volumes/uuid/...
	parts := strings.Split(path, string(filepath.Separator))
	for i, part := range parts {
		if uuidRegex.MatchString(part) {
			if i+1 < len(parts) {
				return part
			}
		}
	}
	return ""
}

func (p *PterodactylAutoSuspend) getServerID(uuid string) (int, error) {
	url := fmt.Sprintf("%s/api/application/servers?filter[uuid]=%s", p.cfg.Plugins.PterodactylAutoSuspend.Hostname, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.cfg.Plugins.PterodactylAutoSuspend.APIKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data struct {
		Data []struct {
			Attributes struct {
				ID int `json:"id"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if len(data.Data) == 0 {
		return 0, fmt.Errorf("no server found for UUID %s", uuid)
	}

	return data.Data[0].Attributes.ID, nil
}

func (p *PterodactylAutoSuspend) suspendServer(serverID int) error {
	url := fmt.Sprintf("%s/api/application/servers/%d/suspend", p.cfg.Plugins.PterodactylAutoSuspend.Hostname, serverID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.cfg.Plugins.PterodactylAutoSuspend.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("suspend API returned status %d", resp.StatusCode)
	}

	logger.Log.Infof("Suspended server ID %d", serverID)
	return nil
}
