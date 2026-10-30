package eis223

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"time"

	"github.com/rinat1313/source_html/internal/apitypes"
)

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
		switch {
		case strings.Contains(local, "purchasenotice"), strings.Contains(local, "purchase_notice"):
			kind = "notice"
		case strings.Contains(local, "purchasecontract"), strings.Contains(local, "contract"):
			kind = "contract"
		}
		if kind == "" {
			continue
		}
		var payload struct {
			XMLName xml.Name
			Body    []byte `xml:",innerxml"`
		}
		payload.XMLName = se.Name
		if err := dec.DecodeElement(&payload, &se); err != nil {
			return nil, err
		}
		frag := append([]byte("<root>"), payload.Body...)
		frag = append(frag, []byte("</root>")...)
		card := parseFragment(kind, frag)
		if card.GUID == "" && card.RegNumber == "" {
			continue
		}
		out = append(out, card)
	}
	return out, nil
}

type bodyFields struct {
	GUID                 string `xml:"guid"`
	RegistrationNumber   string `xml:"registrationNumber"`
	RegistrationNumber2  string `xml:"registration_number"`
	Name                 string `xml:"name"`
	Subject              string `xml:"subject"`
	Description          string `xml:"description"`
	PublicationDateTime  string `xml:"publicationDateTime"`
	CustomerName         string `xml:"customerName"`
	CustomerINN          string `xml:"customerInn"`
}

func parseFragment(kind string, data []byte) apitypes.PurchaseCard {
	var f bodyFields
	_ = xml.Unmarshal(data, &f)
	reg := f.RegistrationNumber
	if reg == "" {
		reg = f.RegistrationNumber2
	}
	title := f.Name
	desc := f.Subject
	if desc == "" {
		desc = f.Description
	}
	dt := parseDate(f.PublicationDateTime)
	return apitypes.PurchaseCard{
		Law:           "223",
		EntityKind:    kind,
		GUID:          strings.TrimSpace(f.GUID),
		RegNumber:     strings.TrimSpace(reg),
		Title:         strings.TrimSpace(title),
		Description:   strings.TrimSpace(desc),
		CustomerName:  strings.TrimSpace(f.CustomerName),
		CustomerINN:   strings.TrimSpace(f.CustomerINN),
		EffectiveDate: dt,
	}
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
