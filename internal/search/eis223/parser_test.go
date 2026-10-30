package eis223

import "testing"

func TestParseNoticeAndContract(t *testing.T) {
	notice := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<data>
  <purchaseNotice>
    <guid>a1b2c3d4-e5f6-7890-abcd-ef1234567890</guid>
    <registrationNumber>223-001</registrationNumber>
    <name>Закупка канцтоваров</name>
    <subject>Бумага и ручки</subject>
    <publicationDateTime>2024-04-01T10:00:00Z</publicationDateTime>
    <customerName>АО Тест</customerName>
    <customerInn>7700112233</customerInn>
  </purchaseNotice>
  <purchaseProtocol>
    <guid>skip-me</guid>
  </purchaseProtocol>
  <purchaseContract>
    <guid>b2c3d4e5-f6a7-8901-bcde-f12345678901</guid>
    <registrationNumber>223-002</registrationNumber>
    <name>Договор поставки</name>
    <description>Канцелярия</description>
    <publicationDateTime>2024-04-15T12:00:00Z</publicationDateTime>
  </purchaseContract>
</data>`)
	cards, err := ParseDocuments(notice)
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].EntityKind != "notice" || cards[0].GUID == "" {
		t.Fatalf("notice card: %+v", cards[0])
	}
	if cards[1].EntityKind != "contract" {
		t.Fatalf("contract card: %+v", cards[1])
	}
}
