package plugins

import (
	"anti-abuse-go/config"
)

type Plugin interface {
	Name() string
	Version() string
	OnStart(cfg *config.Config) error
	OnDetected(path string, matches interface{}) error
	OnScan(path string, content []byte, eventType string) error
}

var registeredPlugins []Plugin

func RegisterPlugin(p Plugin) {
	registeredPlugins = append(registeredPlugins, p)
}

func GetPlugins() []Plugin {
	return registeredPlugins
}

func InitPlugins(cfg *config.Config) error {
	for _, plugin := range registeredPlugins {
		if err := plugin.OnStart(cfg); err != nil {
			return err
		}
	}
	return nil
}