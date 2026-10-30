package ui

import (
	"os"
	"strings"
)

type Config struct {
	Addr               string
	WebRoot            string
	StorageUpstream    string
	SearchUpstream     string
	PublicStorageBase  string
	PublicSearchBase   string
}

func LoadConfig() Config {
	addr := os.Getenv("UI_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	webRoot := os.Getenv("UI_WEB_ROOT")
	if webRoot == "" {
		webRoot = "web"
	}
	publicStorage := os.Getenv("UI_PUBLIC_STORAGE_BASE")
	if publicStorage == "" {
		publicStorage = "/api/storage"
	}
	publicSearch := os.Getenv("UI_PUBLIC_SEARCH_BASE")
	if publicSearch == "" {
		publicSearch = "/api/search"
	}
	storageUpstream := strings.TrimRight(strings.TrimSpace(os.Getenv("STORAGE_URL")), "/")
	if storageUpstream == "" {
		storageUpstream = "http://127.0.0.1:8081"
	}
	searchUpstream := strings.TrimRight(strings.TrimSpace(os.Getenv("SEARCH_URL")), "/")
	if searchUpstream == "" {
		searchUpstream = "http://127.0.0.1:8082"
	}
	return Config{
		Addr:              addr,
		WebRoot:           webRoot,
		StorageUpstream:   storageUpstream,
		SearchUpstream:    searchUpstream,
		PublicStorageBase: strings.TrimRight(publicStorage, "/"),
		PublicSearchBase:  strings.TrimRight(publicSearch, "/"),
	}
}
