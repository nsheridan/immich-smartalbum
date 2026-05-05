package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseConfigConfig_valid(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 30m
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people:
          - Bob
          - Carol
`)
	cfg, err := parseConfig(contents)
	if err != nil {
		t.Fatalf("loadConfig() error: %v", err)
	}
	if cfg.Server != "https://immich.example.com" {
		t.Errorf("server = %q", cfg.Server)
	}
	if cfg.Interval != 30*time.Minute {
		t.Errorf("interval = %v", cfg.Interval)
	}
	if len(cfg.Users) != 1 || cfg.Users[0].Name != "alice" {
		t.Errorf("unexpected users: %+v", cfg.Users)
	}
	if len(cfg.Users[0].Albums[0].People) != 2 {
		t.Errorf("unexpected people: %+v", cfg.Users[0].Albums[0].People)
	}
}

func TestParseConfig_missingAPIKey(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 1h
users:
  - name: alice
    albums:
      - name: Family
        people: [Bob]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for missing api_key, got nil")
	}
}

func TestParseConfig_invalidInterval(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: notaduration
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people: [Bob]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for invalid interval, got nil")
	}
}

func TestParseConfig_albumNoPeople(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 1h
users:
  - name: alice
    api_key: key1
    albums:
      - name: Empty
        people: []
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for album with no people, got nil")
	}
}

func TestParseConfig_missingServer(t *testing.T) {
	contents := strings.NewReader(`
interval: 1h
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people: [Bob]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for missing server, got nil")
	}
}

func TestParseConfig_noUsers(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 1h
users: []
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for empty users, got nil")
	}
}

func TestParseConfig_duplicateUserName(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 1h
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people: [Bob]
  - name: alice
    api_key: key2
    albums:
      - name: Other
        people: [Carol]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for duplicate user name, got nil")
	}
}

func TestParseConfig_invalidLogLevel(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 1h
log_level: trace
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people: [Bob]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for invalid log_level, got nil")
	}
}

func TestParseConfig_intervalTooShort(t *testing.T) {
	contents := strings.NewReader(`
server: https://immich.example.com
interval: 30s
users:
  - name: alice
    api_key: key1
    albums:
      - name: Family
        people: [Bob]
`)
	_, err := parseConfig(contents)
	if err == nil {
		t.Fatal("expected error for interval < 1m, got nil")
	}
}
