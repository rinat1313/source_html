package search

import (
	"os"
	"strings"
)

type Config struct {
	Addr              string
	StorageURL        string
	EIS44SOAPURL      string
	EIS44Sender       string
	EIS44CertFile     string
	EIS44KeyFile      string
	EIS223BaseURL     string
	EIS223Login       string
	EIS223Password    string
	DownloadAllowlist string
}

func LoadConfig() Config {
	addr := os.Getenv("SEARCH_ADDR")
	if addr == "" {
		addr = ":8082"
	}
	return Config{
		Addr:              addr,
		StorageURL:        strings.TrimRight(os.Getenv("STORAGE_URL"), "/"),
		EIS44SOAPURL:      os.Getenv("EIS_44_SOAP_URL"),
		EIS44Sender:       os.Getenv("EIS_SENDER_CODE"),
		EIS44CertFile:     os.Getenv("EIS_44_TLS_CERT"),
		EIS44KeyFile:      os.Getenv("EIS_44_TLS_KEY"),
		EIS223BaseURL:     os.Getenv("EIS_223_BASE_URL"),
		EIS223Login:       os.Getenv("EIS_223_LOGIN"),
		EIS223Password:    os.Getenv("EIS_223_PASSWORD"),
		DownloadAllowlist: os.Getenv("EIS_DOWNLOAD_ALLOWLIST"),
	}
}
