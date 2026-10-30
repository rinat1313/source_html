package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rinat1313/source_html/internal/apitypes"
	"github.com/rinat1313/source_html/internal/logging"
)

type StorageClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewStorageClient(baseURL string) *StorageClient {
	return &StorageClient{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

type storageRun struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message"`
	Law          string  `json:"law"`
	LimitN       int     `json:"limit_n"`
	MatchedSaved int     `json:"matched_saved"`
}

type storageProfile struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func (c *StorageClient) GetProfile(ctx context.Context, id string) (storageProfile, error) {
	var err error
	defer logging.Timed("StorageClient.GetProfile", "id", id)(&err)
	var prof storageProfile
	err = c.getJSON(ctx, "/api/v1/profiles/"+id, &prof)
	return prof, err
}

func (c *StorageClient) CreateRun(ctx context.Context, profileID *string, snapshot json.RawMessage, law string, limitN int) (storageRun, error) {
	var err error
	defer logging.Timed("StorageClient.CreateRun")(&err)
	body, _ := json.Marshal(map[string]any{
		"profile_id":       profileID,
		"config_snapshot":  snapshot,
		"law":              law,
		"limit_n":          limitN,
	})
	var run storageRun
	err = c.postJSON(ctx, "/internal/api/v1/runs", body, &run)
	return run, err
}

func (c *StorageClient) GetRun(ctx context.Context, id string) (storageRun, error) {
	var err error
	defer logging.Timed("StorageClient.GetRun", "id", id)(&err)
	var run storageRun
	err = c.getJSON(ctx, "/internal/api/v1/runs/"+id, &run)
	return run, err
}

func (c *StorageClient) PatchRun(ctx context.Context, id string, patch map[string]any) (storageRun, error) {
	var err error
	defer logging.Timed("StorageClient.PatchRun", "id", id)(&err)
	body, _ := json.Marshal(patch)
	var run storageRun
	err = c.patchJSON(ctx, "/internal/api/v1/runs/"+id, body, &run)
	return run, err
}

func (c *StorageClient) UpsertPurchase(ctx context.Context, runID string, card apitypes.PurchaseCard) (apitypes.UpsertResult, error) {
	var err error
	defer logging.Timed("StorageClient.UpsertPurchase", "run_id", runID)(&err)
	body, _ := json.Marshal(card)
	var res apitypes.UpsertResult
	err = c.postJSON(ctx, "/internal/api/v1/runs/"+runID+"/upsert-purchase", body, &res)
	return res, err
}

func (c *StorageClient) GetRunResults(ctx context.Context, runID string, offset, limit int) (map[string]any, error) {
	var err error
	defer logging.Timed("StorageClient.GetRunResults", "run_id", runID)(&err)
	path := fmt.Sprintf("/api/v1/search/runs/%s/results?offset=%d&limit=%d", runID, offset, limit)
	var out map[string]any
	err = c.getJSON(ctx, path, &out)
	return out, err
}

func (c *StorageClient) getJSON(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage GET %s: %d %s", path, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *StorageClient) postJSON(ctx context.Context, path string, body []byte, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage POST %s: %d %s", path, resp.StatusCode, string(b))
	}
	if dest == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *StorageClient) patchJSON(ctx context.Context, path string, body []byte, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("storage PATCH %s: %d %s", path, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *StorageClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("storage health %d", resp.StatusCode)
	}
	return nil
}
