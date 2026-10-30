package search

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rinat1313/source_html/internal/apitypes"
)

type Server struct {
	Storage      *StorageClient
	Orchestrator *Orchestrator
}

func NewServer(storage *StorageClient, orch *Orchestrator) *Server {
	return &Server{Storage: storage, Orchestrator: orch}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /api/v1/search/runs", s.handleCreateRun)
	mux.HandleFunc("GET /api/v1/search/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /api/v1/search/runs/{id}/results", s.handleRunResults)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.Storage.Ping(r.Context()); err != nil {
		writeErr(w, http.StatusServiceUnavailable, "storage unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createRunRequest struct {
	ProfileID string               `json:"profile_id"`
	Config    *apitypes.SearchConfig `json:"config"`
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	var cfg apitypes.SearchConfig
	var profileID *string
	var snapshot []byte
	if req.Config != nil {
		cfg = *req.Config
		if msg := cfg.Validate(); msg != "" {
			writeErr(w, http.StatusBadRequest, msg)
			return
		}
		snapshot, _ = json.Marshal(cfg)
	} else if req.ProfileID != "" {
		pid := req.ProfileID
		profileID = &pid
		prof, err := s.Storage.GetProfile(r.Context(), req.ProfileID)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "profile not found")
			return
		}
		snapshot = prof.Config
		parsed, err := ParseConfig(prof.Config)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		cfg = parsed
	} else {
		writeErr(w, http.StatusBadRequest, "profile_id or config required")
		return
	}
	run, err := s.Storage.CreateRun(r.Context(), profileID, snapshot, cfg.Law, cfg.Limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "create run failed")
		return
	}
	go s.Orchestrator.RunSearch(context.Background(), run.ID, cfg)
	writeJSON(w, http.StatusAccepted, run)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.Storage.GetRun(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleRunResults(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := s.Storage.GetRunResults(r.Context(), id, offset, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "results failed")
		return
	}
	writeJSON(w, http.StatusOK, out)
}
