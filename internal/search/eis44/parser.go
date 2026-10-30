package eis44

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"time"

	"github.com/rinat1313/source_html/internal/apitypes"
)

type rootDetect struct {
	XMLName xml.Name
	Inner   []byte `xml:",innerxml"`
}

func ParseDocuments(data []byte) ([]apitypes.PurchaseCard, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var out []apitypes.PurchaseCard
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := strings.ToLower(se.Name.Local)
		if strings.Contains(local, "protocol") {
			continue
		}
		kind := ""
		if strings.Contains(local, "contract") && !strings.Contains(local, "procedure") {
			kind = "contract"
		} else if strings.Contains(local, "notification") || strings.HasPrefix(local, "epnotification") {
			kind = "notice"
		} else if local == "contract" {
			kind = "contract"
		}
		if kind == "" {
			continue
		}
		var raw rootDetect
		raw.XMLName = se.Name
		if err := dec.DecodeElement(&raw, &se); err != nil {
			return nil, err
		}
		frag := append([]byte("<root>"), raw.Inner...)
		frag = append(frag, []byte("</root>")...)
		card, err := parseFragment(kind, frag)
		if err != nil {
			continue
		}
		if card.RegNumber == "" {
			continue
		}
		out = append(out, card)
	}
	return out, nil
}

type fields struct {
	RegNumber   string `xml:"purchaseNumber"`
	RegNumber2  string `xml:"regNum"`
	Title       string `xml:"purchaseObjectInfo"`
	Title2      string `xml:"name"`
	Description string `xml:"purchaseSubject"`
	Description2 string `xml:"subject"`
	Customer    string `xml:"fullName"`
	INN         string `xml:"INN"`
	PublishDate string `xml:"publishDate"`
	DocDate     string `xml:"docDate"`
	SignDate    string `xml:"signDate"`
}

func parseFragment(kind string, data []byte) (apitypes.PurchaseCard, error) {
	var f fields
	_ = xml.Unmarshal(data, &f)
	reg := f.RegNumber
	if reg == "" {
		reg = f.RegNumber2
	}
	title := f.Title
	if title == "" {
		title = f.Title2
	}
	desc := f.Description
	if desc == "" {
		desc = f.Description2
	}
	dt := parseDate(firstNonEmpty(f.PublishDate, f.DocDate, f.SignDate))
	return apitypes.PurchaseCard{
		Law:           "44",
		EntityKind:    kind,
		RegNumber:     reg,
		Title:         strings.TrimSpace(title),
		Description:   strings.TrimSpace(desc),
		CustomerName:  strings.TrimSpace(f.Customer),
		CustomerINN:   strings.TrimSpace(f.INN),
		EffectiveDate: dt,
	}, nil
}

func ParseZip(data []byte) ([]apitypes.PurchaseCard, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var all []apitypes.PurchaseCard
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		parsed, err := ParseDocuments(body)
		if err != nil {
			return nil, err
		}
		all = append(all, parsed...)
	}
	return all, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return time.Now().UTC().Format(time.RFC3339)
}

func parseDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC()
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}
