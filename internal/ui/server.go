package ui

import (
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("STORAGE_URL is required")
	}
	if cfg.SearchUpstream == "" {
		return nil, fmt.Errorf("SEARCH_URL is required")
	}
	if st, err := os.Stat(cfg.WebRoot); err != nil || !st.IsDir() {
		return nil, fmt.Errorf("UI web root missing")
	}
	return &Server{cfg: cfg}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /config.json", s.handleConfigJSON)
	mux.HandleFunc("GET /config.js", s.handleConfigJS)
	mux.Handle("/api/storage/", s.rewriteProxy(s.cfg.StorageUpstream, "/api/storage/", "/api/v1/"))
	mux.Handle("/api/search/", s.rewriteProxy(s.cfg.SearchUpstream, "/api/search/", "/api/v1/search/"))
	mux.Handle("/", s.staticHandler())
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleConfigJSON(w http.ResponseWriter, r *http.Request) {
	var err error
	defer logging.Timed("handleConfigJSON")(&err)
	payload := map[string]string{
		"storageAPIBase": s.cfg.PublicStorageBase,
		"searchAPIBase":  s.cfg.PublicSearchBase,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(payload)
}

func (s *Server) handleConfigJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	body := fmt.Sprintf(
		"window.__APP_CONFIG__={storagePrefix:%q,searchPrefix:%q};",
		s.cfg.PublicStorageBase,
		s.cfg.PublicSearchBase,
	)
	_, _ = w.Write([]byte(body))
}

func (s *Server) rewriteProxy(upstream, stripPrefix, apiPrefix string) http.Handler {
	target, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		slog.Error("proxy error", "upstream", upstream, "path", r.URL.Path, "err", e)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
	}
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		suffix := strings.TrimPrefix(req.URL.Path, stripPrefix)
		suffix = strings.TrimPrefix(suffix, "/")
		req.URL.Path = apiPrefix + suffix
		req.Host = target.Host
	}
	return proxy
}

func (s *Server) staticHandler() http.Handler {
	root := http.Dir(s.cfg.WebRoot)
	fileServer := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
