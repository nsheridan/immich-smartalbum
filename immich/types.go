package immich

import "fmt"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type peopleResponse struct {
	People []Person `json:"people"`
}

type asset struct {
	ID string `json:"id"`
}

type searchAssetsResponse struct {
	Assets struct {
		Items    []asset `json:"items"`
		NextPage *string `json:"nextPage"` // nil when no further pages
	} `json:"assets"`
}

type searchMetadataRequest struct {
	PersonIDs []string `json:"personIds,omitempty"`
	AlbumIDs  []string `json:"albumIds,omitempty"`
	Page      int      `json:"page"`
	Size      int      `json:"size"`
}

type Album struct {
	ID        string `json:"id"`
	AlbumName string `json:"albumName"`
}

type createAlbumRequest struct {
	AlbumName string `json:"albumName"`
}

type bulkAssetsRequest struct {
	IDs []string `json:"ids"`
}

type bulkAssetResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("immich API error: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("immich API error: HTTP %d", e.StatusCode)
}
