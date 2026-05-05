package immich

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const userAgent = "immich-smartalbum"

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 100,
				MaxConnsPerHost:     100,
			},
		},
	}
}

func (c *Client) do(method, path string, body io.Reader) ([]byte, error) {
	slog.Debug("request", "method", method, "path", path)
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("user-agent", userAgent)
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil && body != http.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	slog.Debug("response", "method", method, "path", path, "bytes", len(data))
	return data, nil
}

func (c *Client) ListPeople() ([]Person, error) {
	data, err := c.do("GET", "/api/people", nil)
	if err != nil {
		return nil, err
	}
	var resp peopleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return resp.People, nil
}

func (c *Client) SearchAssetsByPerson(personID string) ([]string, error) {
	var ids []string
	page := 1
	for {
		reqBody := searchMetadataRequest{
			PersonIDs: []string{personID},
			Page:      page,
			Size:      1000,
		}
		body := &bytes.Buffer{}
		if err := json.NewEncoder(body).Encode(reqBody); err != nil {
			return nil, err
		}
		data, err := c.do("POST", "/api/search/metadata", body)
		if err != nil {
			return nil, err
		}
		var resp searchAssetsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		slog.Debug("search page", "person_id", personID, "page", page, "items", len(resp.Assets.Items))
		for _, a := range resp.Assets.Items {
			ids = append(ids, a.ID)
		}
		if resp.Assets.NextPage == nil {
			break
		}
		next, err := strconv.Atoi(*resp.Assets.NextPage)
		if err != nil {
			break
		}
		page = next
	}
	return ids, nil
}

func (c *Client) ListAlbums() ([]Album, error) {
	data, err := c.do("GET", "/api/albums?withoutAssets=true", nil)
	if err != nil {
		return nil, err
	}
	var albums []Album
	if err := json.Unmarshal(data, &albums); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	slog.Debug("listed albums", "count", len(albums))
	return albums, nil
}

func (c *Client) GetAlbumAssetIDs(albumID string) (map[string]struct{}, error) {
	data, err := c.do("GET", "/api/albums/"+albumID, nil)
	if err != nil {
		return nil, err
	}
	var resp albumDetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	ids := make(map[string]struct{}, len(resp.Assets))
	for _, a := range resp.Assets {
		ids[a.ID] = struct{}{}
	}
	return ids, nil
}

func (c *Client) AddAssetsToAlbum(albumID string, assetIDs []string) error {
	reqBody := bulkAssetsRequest{IDs: assetIDs}
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(reqBody); err != nil {
		return err
	}
	data, err := c.do("PUT", "/api/albums/"+albumID+"/assets", body)
	if err != nil {
		return err
	}
	var results []bulkAssetResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	var failed []string
	for _, r := range results {
		if !r.Success {
			failed = append(failed, r.ID)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to add %d asset(s) to album: %v", len(failed), failed)
	}
	return nil
}

func AuthError(err error) bool {
	e, ok := errors.AsType[*APIError](err)
	if ok {
		return e.StatusCode == http.StatusUnauthorized
	}
	return false
}
