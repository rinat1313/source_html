package eis223

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/rinat1313/source_html/internal/apitypes"
	"github.com/rinat1313/source_html/internal/logging"
)

type Client interface {
	Collect(ctx context.Context, params apitypes.EIS223Params) ([]apitypes.PurchaseCard, error)
}

type RealClient struct {
	HTTP     *http.Client
	BaseURL  string
	Login    string
	Password string
}

func (c *RealClient) Collect(ctx context.Context, params apitypes.EIS223Params) ([]apitypes.PurchaseCard, error) {
	var err error
	defer logging.Timed("eis223.Collect")(&err)
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 90 * time.Second}
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		err = fmt.Errorf("EIS 223 base URL not configured")
		return nil, err
	}
	types := params.EntityTypes
	if len(types) == 0 {
		types = []string{"purchaseNotice", "purchaseContract"}
	}
	var all []apitypes.PurchaseCard
	for _, entity := range types {
		cards, cerr := c.fetchEntity(ctx, base, entity, params)
		if cerr != nil {
			err = cerr
			return nil, err
		}
		all = append(all, cards...)
	}
	return all, nil
}

func (c *RealClient) fetchEntity(ctx context.Context, base, entity string, params apitypes.EIS223Params) ([]apitypes.PurchaseCard, error) {
	url := base + "/223/integration/rest/1.3/publishedInfo/" + entity
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("login", c.Login)
	_ = w.WriteField("password", c.Password)
	if params.Period.From != "" {
		_ = w.WriteField("startDate", params.Period.From)
	}
	if params.Period.To != "" {
		_ = w.WriteField("endDate", params.Period.To)
	}
	if params.Region != "" {
		_ = w.WriteField("region", params.Region)
	}
	_ = w.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("223 http status %d", resp.StatusCode)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	if data[0] != '<' {
		return nil, fmt.Errorf("unexpected 223 response")
	}
	var probe struct {
		XMLName xml.Name
	}
	if xml.Unmarshal(data, &probe) != nil {
		return nil, fmt.Errorf("invalid 223 xml")
	}
	return ParseDocuments(data)
}
