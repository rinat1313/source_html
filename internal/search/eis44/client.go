package eis44

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rinat1313/source_html/internal/apitypes"
	"github.com/rinat1313/source_html/internal/logging"
	"github.com/rinat1313/source_html/internal/search/eishttp"
)

type Client interface {
	Collect(ctx context.Context, params apitypes.EIS44Params) ([]apitypes.PurchaseCard, error)
}

type RealClient struct {
	HTTP       *http.Client
	SOAPURL    string
	SenderCode string
	Downloader *eishttp.Downloader
}

type envelopeResp struct {
	Body struct {
		Response struct {
			ArchiveURLs []string `xml:"archiveUrl"`
			NoData      bool     `xml:"noData"`
			RequestUID  string   `xml:"requestUid"`
		} `xml:"getPublicDocsByOrgRegionResponse"`
		Prepared struct {
			Ready       bool     `xml:"ready"`
			ArchiveURLs []string `xml:"archiveUrl"`
		} `xml:"getPreparedPartResponse"`
		Fault *struct {
			String string `xml:"faultstring"`
		} `xml:"Fault"`
	} `xml:"Body"`
}

func (c *RealClient) Collect(ctx context.Context, params apitypes.EIS44Params) ([]apitypes.PurchaseCard, error) {
	var err error
	defer logging.Timed("eis44.Collect")(&err)
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 90 * time.Second}
	}
	if strings.TrimSpace(c.SOAPURL) == "" {
		err = fmt.Errorf("EIS 44 SOAP URL not configured")
		return nil, err
	}
	region := params.OrgRegion
	if region == "" {
		region = "77000000000"
	}
	docTypes := params.DocumentTypes
	if len(docTypes) == 0 {
		docTypes = []string{"epNotification", "contract"}
	}
	reqBody := buildOrgRegionRequest(c.SenderCode, region, docTypes, params.Period)
	respBody, err := c.soapCall(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	var env envelopeResp
	if err = xml.Unmarshal(respBody, &env); err != nil {
		return nil, err
	}
	if env.Body.Fault != nil {
		err = fmt.Errorf("soap fault: %s", env.Body.Fault.String)
		return nil, err
	}
	urls := env.Body.Response.ArchiveURLs
	requestUID := env.Body.Response.RequestUID
	if len(urls) == 0 && requestUID != "" {
		urls, err = c.pollPrepared(ctx, requestUID)
		if err != nil {
			return nil, err
		}
	}
	if len(urls) == 0 {
		return nil, nil
	}
	var cards []apitypes.PurchaseCard
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		zipData, derr := c.Downloader.GET(u)
		if derr != nil {
			err = derr
			return nil, err
		}
		parsed, perr := ParseZip(zipData)
		if perr != nil {
			err = perr
			return nil, err
		}
		cards = append(cards, parsed...)
	}
	return cards, nil
}

func (c *RealClient) pollPrepared(ctx context.Context, requestUID string) ([]string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
		body := buildPreparedPartRequest(c.SenderCode, requestUID)
		respBody, err := c.soapCall(ctx, body)
		if err != nil {
			return nil, err
		}
		var env envelopeResp
		if err = xml.Unmarshal(respBody, &env); err != nil {
			return nil, err
		}
		if env.Body.Prepared.Ready && len(env.Body.Prepared.ArchiveURLs) > 0 {
			return env.Body.Prepared.ArchiveURLs, nil
		}
	}
	return nil, fmt.Errorf("archive not ready")
}

func (c *RealClient) soapCall(ctx context.Context, envelope string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.SOAPURL, strings.NewReader(envelope))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
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
		return nil, fmt.Errorf("soap http status %d", resp.StatusCode)
	}
	return data, nil
}

func buildOrgRegionRequest(sender, region string, docTypes []string, period apitypes.Period) string {
	typesXML := ""
	for _, dt := range docTypes {
		typesXML += fmt.Sprintf("<documentType44>%s</documentType44>", xmlEscape(dt))
	}
	periodXML := ""
	if period.From != "" {
		periodXML += fmt.Sprintf("<dateFrom>%s</dateFrom>", xmlEscape(period.From))
	}
	if period.To != "" {
		periodXML += fmt.Sprintf("<dateTo>%s</dateTo>", xmlEscape(period.To))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <getPublicDocsByOrgRegionRequest>
      <index><sender>%s</sender></index>
      <selectionParams>
        <orgRegion>%s</orgRegion>
        %s
        %s
      </selectionParams>
    </getPublicDocsByOrgRegionRequest>
  </soapenv:Body>
</soapenv:Envelope>`, xmlEscape(sender), xmlEscape(region), typesXML, periodXML)
}

func buildPreparedPartRequest(sender, requestUID string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <getPreparedPartRequest>
      <index><sender>%s</sender></index>
      <requestUid>%s</requestUid>
    </getPreparedPartRequest>
  </soapenv:Body>
</soapenv:Envelope>`, xmlEscape(sender), xmlEscape(requestUID))
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}
