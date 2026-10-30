package ui

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/rinat1313/source_html/internal/logging"
)

type Server struct {
	cfg Config
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.StorageUpstream == "" {
		return nil, errors.New("STORAGE_URL is required")
	}
	if cfg.SearchUpstream == "" {
		return nil, errors.New("SEARCH_URL is required")
	}
	if st, err := os.Stat(cfg.WebRoot); err != nil || !st.IsDir() {
		return nil, errors.New("UI web root missing")
	}
	return &Server{cfg: cfg}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /config.json", s.handleConfig)
	mux.Handle("/api/storage/", s.proxyHandler(s.cfg.StorageUpstream, "/api/storage"))
	mux.Handle("/api/search/", s.proxyHandler(s.cfg.SearchUpstream, "/api/search"))
	mux.Handle("/", s.staticHandler())
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	var err error
	defer logging.Timed("handleConfig")(&err)
	payload := map[string]string{
		"storageAPIBase": s.cfg.PublicStorageBase,
		"searchAPIBase":  s.cfg.PublicSearchBase,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(payload)
}

func (s *Server) proxyHandler(upstream string, mountPrefix string) http.Handler {
	target, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		slog.Error("proxy error", "upstream", upstream, "path", r.URL.Path, "err", e)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	prefix := strings.TrimRight(mountPrefix, "/") + "/"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})
}

func (s *Server) staticHandler() http.Handler {
	return http.FileServer(http.Dir(s.cfg.WebRoot))
}
