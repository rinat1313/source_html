package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rinat1313/source_html/internal/logging"
	"github.com/rinat1313/source_html/internal/search"
	"github.com/rinat1313/source_html/internal/search/eis44"
	"github.com/rinat1313/source_html/internal/search/eis223"
	"github.com/rinat1313/source_html/internal/search/eishttp"
)

func main() {
	logging.Setup("search")
	cfg := search.LoadConfig()
	if cfg.StorageURL == "" {
		slog.Error("STORAGE_URL is required")
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 120 * time.Second}
	if cfg.EIS44CertFile != "" && cfg.EIS44KeyFile != "" {
		if _, errCert := os.Stat(cfg.EIS44CertFile); errCert != nil {
			slog.Warn("EIS_44_TLS_CERT not found, mTLS skipped", "path", cfg.EIS44CertFile, "err", errCert)
		} else if _, errKey := os.Stat(cfg.EIS44KeyFile); errKey != nil {
			slog.Warn("EIS_44_TLS_KEY not found, mTLS skipped", "path", cfg.EIS44KeyFile, "err", errKey)
		} else {
			cert, err := tls.LoadX509KeyPair(cfg.EIS44CertFile, cfg.EIS44KeyFile)
			if err != nil {
				slog.Warn("tls cert load failed, mTLS skipped", "err", err)
			} else {
				httpClient.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{
						Certificates: []tls.Certificate{cert},
						MinVersion:   tls.VersionTLS12,
					},
				}
			}
		}
	}

	storage := search.NewStorageClient(cfg.StorageURL)
	client44 := &eis44.RealClient{
		HTTP:       httpClient,
		SOAPURL:    cfg.EIS44SOAPURL,
		SenderCode: cfg.EIS44Sender,
		Downloader: eishttp.NewDownloader(httpClient, cfg.DownloadAllowlist),
	}
	client223 := &eis223.RealClient{
		HTTP:     httpClient,
		BaseURL:  cfg.EIS223BaseURL,
		Login:    cfg.EIS223Login,
		Password: cfg.EIS223Password,
	}
	orch := &search.Orchestrator{
		Storage:   storage,
		Client44:  client44,
		Client223: client223,
	}
	srv := search.NewServer(storage, orch)
	httpSrv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("search listening", "addr", cfg.Addr)
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
