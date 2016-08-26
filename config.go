package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl"
)

const LocalMode = "local"
const GlobalMode = "global"

type Config struct {
	ConsulAddress   string `hcl:"consul_address"`
	ConsulToken     string `hcl:"token"`
	DevMode         bool   `hcl:"dev_mode"`
	NodeWatch       string `hcl:"node_watch"`
	ServiceWatch    string `hcl:"service_watch"`
	ChangeThreshold int    `hcl:"change_threshold"`

	LogLevel string `hcl:"log_level"`

	Services []ServiceConfig `hcl:"service"`
	Handlers HandlerConfig   `hcl:"handlers"`
}

type ServiceConfig struct {
	Name            string   `hcl:",key"`
	ChangeThreshold int      `hcl:"change_threshold"`
	DistinctTags    bool     `hcl:"distinct_tags"`
	IgnoredTags     []string `hcl:"ignored_tags"`
}

type HandlerConfig struct {
	StdoutHandlers    []StdoutHandler    `hcl:"stdout"`
	EmailHandlers     []EmailHandler     `hcl:"email"`
	PagerdutyHandlers []PagerdutyHandler `hcl:"pagerduty"`
}

// Parses a given file path for config and returns a Config object and an array
// of AlertHandlers
func ParseConfigFile(path string) (*Config, []AlertHandler, error) {
	// Read the file contents
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("Error loading config file: %s", err)
	}
	raw := string(bytes)

	return ParseConfig(raw)
}

func DefaultConfig() (*Config, []AlertHandler) {
	config, handlers, _ := ParseConfig("{}")
	return config, handlers
}

// Parses the given config string and returns a Config object and an array
// of AlertHandlers
func ParseConfig(raw string) (*Config, []AlertHandler, error) {
	config := &Config{}

	if err := hcl.Decode(&config, raw); err != nil {
		return nil, nil, err
	}

	// Set default global config
	if config.ConsulAddress == "" {
		config.ConsulAddress = "localhost:8500"
	}

	if config.ChangeThreshold == 0 {
		config.ChangeThreshold = 60
	}

	if config.LogLevel == "" {
		config.LogLevel = "INFO"
	}

	validWatchModes := []string{LocalMode, GlobalMode}

	if config.NodeWatch == "" {
		config.NodeWatch = "local"
	} else if !contains(validWatchModes, config.NodeWatch) {
		return nil, nil, fmt.Errorf("Unrecognized node_watch setting: %s", config.NodeWatch)
	}

	if config.ServiceWatch == "" {
		config.ServiceWatch = "local"
	} else if !contains(validWatchModes, config.ServiceWatch) {
		return nil, nil, fmt.Errorf("Unrecognized service_watch setting: %s", config.ServiceWatch)
	}

	// Set default service config
	for _, service := range config.Services {
		if service.ChangeThreshold == 0 {
			service.ChangeThreshold = config.ChangeThreshold
		}
	}

	// Configure alert handlers
	handlers := make([]AlertHandler, 0)

	for _, handler := range config.Handlers.StdoutHandlers {
		if handler.LogLevel == "" {
			handler.LogLevel = "warn"
		}
		_, err := log.ParseLevel(handler.LogLevel)
		if err != nil {
			return nil, nil, fmt.Errorf("Error parsing loglevel %s: %s", handler.LogLevel, err)
		}
		log.Infof("Handler stdout enabled with loglevel %s", handler.LogLevel)
		handlers = append(handlers, handler)
	}

	for _, handler := range config.Handlers.EmailHandlers {
		log.Infof("Handler email enabled with recipients: %v", handler.Recipients)
		handlers = append(handlers, handler)
	}

	for _, handler := range config.Handlers.PagerdutyHandlers {
		log.Infof("Handler pagerduty enabled")
		handlers = append(handlers, handler)
	}

	return config, handlers, nil
}

func (config *Config) getServiceConfig(name string) *ServiceConfig {
	for _, service := range config.Services {
		if service.Name == name {
			return &service
		}
	}
	return nil
}
