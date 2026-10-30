package eis44

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNoticeAndContract(t *testing.T) {
	notice := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<export>
  <epNotificationEF2020>
    <purchaseNumber>0123456789012345678901234</purchaseNumber>
    <purchaseObjectInfo>Поставка бумаги офисной</purchaseObjectInfo>
    <purchaseSubject>Бумага для принтеров</purchaseSubject>
    <publishDate>2024-05-10T12:00:00</publishDate>
    <customer>
      <fullName>ГБУ Тест</fullName>
      <INN>7700000000</INN>
    </customer>
  </epNotificationEF2020>
</export>`)
	contract := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<export>
  <contract>
    <regNum>0123456789012345678901234</regNum>
    <name>Контракт на поставку бумаги</name>
    <subject>Бумага офисная</subject>
    <signDate>2024-06-01</signDate>
    <customer>
      <fullName>ГБУ Тест</fullName>
    </customer>
  </contract>
</export>`)
	nCards, err := ParseDocuments(notice)
	if err != nil {
		t.Fatal(err)
	}
	if len(nCards) != 1 || nCards[0].EntityKind != "notice" {
		t.Fatalf("notice: %+v", nCards)
	}
	if nCards[0].Title == "" || nCards[0].RegNumber == "" {
		t.Fatalf("notice fields: %+v", nCards[0])
	}
	cCards, err := ParseDocuments(contract)
	if err != nil {
		t.Fatal(err)
	}
	if len(cCards) != 1 || cCards[0].EntityKind != "contract" {
		t.Fatalf("contract: %+v", cCards)
	}
}

func TestParseFixtureFiles(t *testing.T) {
	dir := filepath.Join("testdata")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skip("no testdata")
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".xml" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseDocuments(body); err != nil {
			t.Fatalf("file %s: %v", e.Name(), err)
		}
	}
}
