package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	AppName           = "classreach"
	EnvPrefix         = "CLASSREACH"
	DefaultBaseURL    = "https://providencewilmington.classreach.com"
	DefaultOriginHost = "classreach.azurewebsites.net"
	ConfigFilename    = "config.yaml"
)

type Config struct {
	BaseURL    string `yaml:"base_url"`
	OriginHost string `yaml:"origin_host"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

func Default() Config {
	return Config{
		BaseURL:    DefaultBaseURL,
		OriginHost: DefaultOriginHost,
	}
}

func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, AppName, ConfigFilename)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ConfigFilename)
	}
	return filepath.Join(home, ".config", AppName, ConfigFilename)
}

func Load(path string) (*Config, error) {
	cfg := Default()
	if path == "" {
		path = DefaultPath()
	}
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config file %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	applyEnv(&cfg)
	normalize(&cfg)
	return &cfg, nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath()
	}
	normalize(&cfg)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required; set --base-url or config file base_url")
	}
	if c.OriginHost == "" {
		return fmt.Errorf("origin host is required; set --origin-host or config file origin_host")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required; set config file username")
	}
	if c.Password == "" {
		return fmt.Errorf("password is required; set config file password")
	}
	return nil
}

func (c Config) Redacted() map[string]string {
	return map[string]string{
		"base_url":    c.BaseURL,
		"origin_host": c.OriginHost,
		"username":    redact(c.Username),
		"password":    redact(c.Password),
	}
}

func applyEnv(cfg *Config) {
	if value := os.Getenv(EnvPrefix + "_BASE_URL"); value != "" {
		cfg.BaseURL = value
	}
	if value := os.Getenv(EnvPrefix + "_ORIGIN_HOST"); value != "" {
		cfg.OriginHost = value
	}
}

func normalize(cfg *Config) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.OriginHost = strings.TrimSpace(cfg.OriginHost)
	cfg.Username = strings.TrimSpace(cfg.Username)
}

func redact(value string) string {
	if value == "" {
		return ""
	}
	return "redacted"
}
