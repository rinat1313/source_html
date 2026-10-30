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
	return Config{
		Addr:              addr,
		WebRoot:           webRoot,
		StorageUpstream:   strings.TrimRight(os.Getenv("STORAGE_URL"), "/"),
		SearchUpstream:    strings.TrimRight(os.Getenv("SEARCH_URL"), "/"),
		PublicStorageBase: strings.TrimRight(publicStorage, "/"),
		PublicSearchBase:  strings.TrimRight(publicSearch, "/"),
	}
}
