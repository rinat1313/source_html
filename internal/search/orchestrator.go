package search

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"

	"github.com/rinat1313/source_html/internal/apitypes"
	"github.com/rinat1313/source_html/internal/logging"
	"github.com/rinat1313/source_html/internal/search/eis44"
	"github.com/rinat1313/source_html/internal/search/eis223"
	"github.com/rinat1313/source_html/internal/search/keywords"
)

const eisUnavailableMsg = "ЕИС недоступна"

type Orchestrator struct {
	Storage *StorageClient
	Client44 eis44.Client
	Client223 eis223.Client
}

func (o *Orchestrator) RunSearch(ctx context.Context, runID string, cfg apitypes.SearchConfig) {
	var err error
	defer logging.Timed("RunSearch", "run_id", runID)(&err)
	running := "running"
	_, _ = o.Storage.PatchRun(ctx, runID, map[string]any{"status": running})

	laws := []string{cfg.Law}
	if cfg.Law == "both" {
		laws = []string{"44", "223"}
	}

	eisFailed := false
	var eisErr error

	for _, law := range laws {
		run, gerr := o.Storage.GetRun(ctx, runID)
		if gerr != nil {
			return
		}
		if run.MatchedSaved >= cfg.Limit {
			break
		}
		var cards []apitypes.PurchaseCard
		var cerr error
		switch law {
		case "44":
			cards, cerr = o.Client44.Collect(ctx, cfg.EIS44)
		case "223":
			cards, cerr = o.Client223.Collect(ctx, cfg.EIS223)
		}
		if cerr != nil {
			if isTransportError(cerr) {
				eisFailed = true
				eisErr = cerr
			}
			continue
		}
		for _, card := range cards {
			run, gerr = o.Storage.GetRun(ctx, runID)
			if gerr != nil || run.MatchedSaved >= cfg.Limit {
				break
			}
			if !keywords.Match(card.Title, card.Description, cfg.Keywords) {
				continue
			}
			_, uerr := o.Storage.UpsertPurchase(ctx, runID, card)
			if uerr != nil {
				continue
			}
		}
	}

	finished := true
	if eisFailed {
		msg := eisUnavailableMsg
		status := "failed"
		_, _ = o.Storage.PatchRun(ctx, runID, map[string]any{
			"status":        status,
			"error_message": msg,
			"finished":      finished,
		})
		_ = eisErr
		return
	}
	status := "completed"
	_, _ = o.Storage.PatchRun(ctx, runID, map[string]any{
		"status":   status,
		"finished": finished,
	})
}

func isTransportError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, sub := range []string{"tls", "connection refused", "timeout", "no such host", "soap http", "223 http"} {
		if strings.Contains(msg, sub) {
			return true
		}
	}
	return false
}

func ParseConfig(raw json.RawMessage) (apitypes.SearchConfig, error) {
	var cfg apitypes.SearchConfig
	if len(raw) == 0 {
		return cfg, errors.New("empty config")
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	if msg := cfg.Validate(); msg != "" {
		return cfg, errors.New(msg)
	}
	return cfg, nil
}
