package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rinat1313/source_html/internal/apitypes"
)

type pingPool interface {
	Ping(ctx context.Context) error
}

type Server struct {
	repo *Repo
	pool pingPool
}

func NewServer(repo *Repo, pool pingPool) *Server {
	return &Server{repo: repo, pool: pool}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /api/v1/profiles", s.handleListProfiles)
	mux.HandleFunc("POST /api/v1/profiles", s.handleCreateProfile)
	mux.HandleFunc("GET /api/v1/profiles/{id}", s.handleGetProfile)
	mux.HandleFunc("PUT /api/v1/profiles/{id}", s.handleUpdateProfile)
	mux.HandleFunc("DELETE /api/v1/profiles/{id}", s.handleDeleteProfile)
	mux.HandleFunc("GET /api/v1/purchases", s.handleListPurchases)
	mux.HandleFunc("GET /api/v1/purchases/{id}", s.handleGetPurchase)
	mux.HandleFunc("GET /api/v1/search/runs/{id}/results", s.handleRunResults)
	mux.HandleFunc("POST /internal/api/v1/runs", s.handleInternalCreateRun)
	mux.HandleFunc("GET /internal/api/v1/runs/{id}", s.handleInternalGetRun)
	mux.HandleFunc("PATCH /internal/api/v1/runs/{id}", s.handleInternalPatchRun)
	mux.HandleFunc("POST /internal/api/v1/runs/{id}/upsert-purchase", s.handleInternalUpsertPurchase)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createProfileBody struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var body createProfileBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeErr(w, http.StatusBadRequest, "name required")
		return
	}
	if len(body.Config) > 0 {
		var cfg apitypes.SearchConfig
		if err := json.Unmarshal(body.Config, &cfg); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid config")
			return
		}
		if msg := cfg.Validate(); msg != "" {
			writeErr(w, http.StatusBadRequest, msg)
			return
		}
	}
	p, err := s.repo.CreateProfile(r.Context(), strings.TrimSpace(body.Name), body.Config)
	if err != nil {
		slog.Error("CreateProfile failed", "err", err)
		writeErr(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	list, err := s.repo.ListProfiles(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list failed")
		return
	}
	if list == nil {
		list = []Profile{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.repo.GetProfile(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "get failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body createProfileBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(body.Config) > 0 {
		var cfg apitypes.SearchConfig
		if err := json.Unmarshal(body.Config, &cfg); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid config")
			return
		}
		if msg := cfg.Validate(); msg != "" {
			writeErr(w, http.StatusBadRequest, msg)
			return
		}
	}
	p, err := s.repo.UpdateProfile(r.Context(), id, strings.TrimSpace(body.Name), body.Config)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.repo.DeleteProfile(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListPurchases(w http.ResponseWriter, r *http.Request) {
	law := r.URL.Query().Get("law")
	q := r.URL.Query().Get("q")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, total, err := s.repo.ListPurchases(r.Context(), law, q, offset, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "offset": offset, "limit": limit})
}

func (s *Server) handleGetPurchase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.repo.GetPurchase(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "get failed")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleRunResults(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, total, err := s.repo.ListRunResults(r.Context(), runID, offset, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "offset": offset, "limit": limit})
}

type internalCreateRunBody struct {
	ProfileID      *string         `json:"profile_id"`
	ConfigSnapshot json.RawMessage `json:"config_snapshot"`
	Law            string          `json:"law"`
	LimitN         int             `json:"limit_n"`
}

func (s *Server) handleInternalCreateRun(w http.ResponseWriter, r *http.Request) {
	var body internalCreateRunBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	run, err := s.repo.CreateRun(r.Context(), body.ProfileID, body.ConfigSnapshot, body.Law, body.LimitN)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (s *Server) handleInternalGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.repo.GetRun(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "get failed")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

type internalPatchRunBody struct {
	Status       *string `json:"status"`
	ErrorMessage *string `json:"error_message"`
	MatchedSaved *int    `json:"matched_saved"`
	Finished     *bool   `json:"finished"`
}

func (s *Server) handleInternalPatchRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body internalPatchRunBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	patch := RunPatch{
		Status:       body.Status,
		ErrorMessage: body.ErrorMessage,
		MatchedSaved: body.MatchedSaved,
	}
	if body.Finished != nil && *body.Finished {
		now := time.Now().UTC()
		patch.FinishedAt = &now
	}
	run, err := s.repo.PatchRun(r.Context(), id, patch)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "patch failed")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleInternalUpsertPurchase(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	var card apitypes.PurchaseCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := s.repo.UpsertPurchaseForRun(r.Context(), runID, card)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
