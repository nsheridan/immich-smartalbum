package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"go.yaml.in/yaml/v4"
)

type config struct {
	Server   string        `yaml:"server"`
	Interval time.Duration `yaml:"interval"`
	LogLevel string        `yaml:"log_level"`
	Users    []userConfig  `yaml:"users"`
}

type userConfig struct {
	Name   string        `yaml:"name"`
	APIKey string        `yaml:"api_key"`
	Albums []albumConfig `yaml:"albums"`
}

type albumConfig struct {
	Name    string   `yaml:"name"`
	AlbumID string   `yaml:"album_id"`
	People  []string `yaml:"people"`
}

func loadConfig(path string) (*config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return parseConfig(fd)
}

func parseConfig(r io.Reader) (*config, error) {
	var cfg config
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Interval < time.Minute {
		return nil, fmt.Errorf("interval %q is too short (minimum 1m)", cfg.Interval)
	}
	if cfg.Server == "" {
		return nil, fmt.Errorf("server is required")
	}
	if len(cfg.Users) == 0 {
		return nil, fmt.Errorf("no users configured")
	}

	seen := make(map[string]struct{}, len(cfg.Users))
	for _, u := range cfg.Users {
		if _, dup := seen[u.Name]; dup {
			return nil, fmt.Errorf("duplicate user name %q", u.Name)
		}
		seen[u.Name] = struct{}{}
		if u.APIKey == "" {
			return nil, fmt.Errorf("user %q has no api_key", u.Name)
		}
		for _, a := range u.Albums {
			if len(a.People) == 0 {
				return nil, fmt.Errorf("user %q album %q has no people", u.Name, a.Name)
			}
		}
	}

	switch cfg.LogLevel {
	case "", "info", "debug":
	default:
		return nil, fmt.Errorf("log_level %q invalid (must be info or debug)", cfg.LogLevel)
	}

	return &cfg, nil
}
