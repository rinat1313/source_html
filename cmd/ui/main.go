package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rinat1313/source_html/internal/logging"
	"github.com/rinat1313/source_html/internal/ui"
)

func main() {
	logging.Setup("ui")
	cfg := ui.LoadConfig()
	srv, err := ui.NewServer(cfg)
	if err != nil {
		slog.Error("ui setup failed", "err", err)
		os.Exit(1)
	}
	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		slog.Info("ui listening", "addr", cfg.Addr, "storage_upstream", cfg.StorageUpstream, "search_upstream", cfg.SearchUpstream)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen failed", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shCtx)
}
