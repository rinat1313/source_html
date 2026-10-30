package eishttp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Downloader struct {
	Client    *http.Client
	Allowlist []string
}

func NewDownloader(client *http.Client, allowlistCSV string) *Downloader {
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	var hosts []string
	for _, h := range strings.Split(allowlistCSV, ",") {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			hosts = append(hosts, h)
		}
	}
	if len(hosts) == 0 {
		hosts = []string{
			"int.zakupki.gov.ru",
			"int44-ttls-cert.zakupki.gov.ru",
			"int44.zakupki.gov.ru",
			"int223.zakupki.gov.ru",
		}
	}
	return &Downloader{Client: client, Allowlist: hosts}
}

func (d *Downloader) GET(rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	host := strings.ToLower(u.Hostname())
	ok := false
	for _, a := range d.Allowlist {
		if host == a || strings.HasSuffix(host, "."+a) {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("download host not allowed: %s", host)
	}
	resp, err := d.Client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
