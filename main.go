package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"nsheridan.dev/immich-smartalbum/immich"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	clients := make([]*immich.Client, len(cfg.Users))
	for i, u := range cfg.Users {
		clients[i] = immich.New(cfg.Server, u.APIKey)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	run(cfg, clients)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			run(cfg, clients)
		case <-sig:
			slog.Info("signal received, shutting down")
			return
		}
	}
}

func run(cfg *config, clients []*immich.Client) {
	slog.Info("starting scan")
	results := make([]userResult, len(cfg.Users))
	var wg sync.WaitGroup
	for i, user := range cfg.Users {
		i, user, client := i, user, clients[i]
		wg.Go(func() {
			results[i] = processUser(user, client)
		})
	}
	wg.Wait()

	for _, r := range results {
		if r.Err != nil {
			if immich.AuthError(r.Err) {
				slog.Warn("authentication failed, skipping user", "user", r.Name)
			} else {
				slog.Error("user scan failed", "user", r.Name, "err", r.Err)
			}
		}
	}
	printSummary(results)
	slog.Info("scan complete", "next_run", cfg.Interval)
}

func printSummary(results []userResult) {
	for _, ur := range results {
		if ur.Err != nil {
			slog.Error("user scan failed", "user", ur.Name, "err", ur.Err)
			continue
		}
		for _, ar := range ur.Albums {
			for _, w := range ar.Warnings {
				slog.Warn(w, "user", ur.Name)
			}
			if ar.Err != nil {
				slog.Error("album scan failed", "user", ur.Name, "album", ar.Name, "err", ar.Err)
				continue
			}
			if ar.Added == 0 {
				slog.Info("nothing new", "user", ur.Name, "album", ar.Name)
			} else {
				slog.Info("assets added", "user", ur.Name, "album", ar.Name, "count", ar.Added)
			}
		}
	}
}
