package main

import (
	"fmt"
	"log/slog"

	"nsheridan.dev/immich-smartalbum/immich"
)

type albumResult struct {
	Name     string
	Added    int
	Warnings []string
	Err      error
}

type userResult struct {
	Name   string
	Albums []albumResult
	Err    error
}

func processUser(cfg userConfig, client *immich.Client) userResult {
	result := userResult{Name: cfg.Name}

	people, err := client.ListPeople()
	if err != nil {
		result.Err = fmt.Errorf("list people: %w", err)
		return result
	}
	slog.Debug("listed people", "user", cfg.Name, "count", len(people))

	nameToID := make(map[string]string, len(people))
	for _, p := range people {
		nameToID[p.Name] = p.ID
		if p.Name != "" {
			slog.Debug("person", "name", p.Name, "id", p.ID)
		}
	}

	for _, albumCfg := range cfg.Albums {
		ar := processAlbum(albumCfg, nameToID, client)
		result.Albums = append(result.Albums, ar)
	}
	return result
}

func processAlbum(cfg albumConfig, nameToID map[string]string, client *immich.Client) albumResult {
	ar := albumResult{Name: cfg.Name}

	var personIDs []string
	for _, name := range cfg.People {
		id, ok := nameToID[name]
		if !ok {
			ar.Warnings = append(ar.Warnings, fmt.Sprintf("person %q not found in Immich, skipping", name))
			continue
		}
		slog.Debug("resolved person", "album", cfg.Name, "person", name, "id", id)
		personIDs = append(personIDs, id)
	}

	albumID := cfg.AlbumID
	if albumID != "" {
		slog.Debug("using configured album ID", "album", cfg.Name, "id", albumID)
	} else {
		var err error
		albumID, err = findAlbum(cfg.Name, client)
		if err != nil {
			ar.Err = fmt.Errorf("find album %q: %w", cfg.Name, err)
			return ar
		}
	}

	// Fetch existing album asset IDs. GET /api/albums/{id} returns all assets in
	// one response with no pagination - should be ok for most album sizes.
	existing, err := client.GetAlbumAssetIDs(albumID)
	if err != nil {
		ar.Err = fmt.Errorf("get album assets: %w", err)
		return ar
	}

	candidates := make(map[string]struct{})
	for _, pid := range personIDs {
		ids, err := client.SearchAssetsByPerson(pid)
		if err != nil {
			w := fmt.Sprintf("search assets for person ID %q: %v - skipping", pid, err)
			ar.Warnings = append(ar.Warnings, w)
			continue
		}
		slog.Debug("assets found for person", "person_id", pid, "count", len(ids))
		for _, id := range ids {
			candidates[id] = struct{}{}
		}
	}
	slog.Debug("candidate assets", "album", cfg.Name, "count", len(candidates))

	var toAdd []string
	for id := range candidates {
		if _, ok := existing[id]; !ok {
			toAdd = append(toAdd, id)
		}
	}
	slog.Debug("assets to add", "album", cfg.Name, "count", len(toAdd))

	if len(toAdd) == 0 {
		return ar
	}

	if err := client.AddAssetsToAlbum(albumID, toAdd); err != nil {
		ar.Err = fmt.Errorf("add assets to album %q: %w", cfg.Name, err)
		return ar
	}
	ar.Added = len(toAdd)
	return ar
}

func findAlbum(name string, client *immich.Client) (string, error) {
	albums, err := client.ListAlbums()
	if err != nil {
		return "", err
	}
	for _, a := range albums {
		if a.AlbumName == name {
			slog.Debug("album found", "name", name, "id", a.ID)
			return a.ID, nil
		}
	}
	return "", fmt.Errorf("album %q not found", name)
}
