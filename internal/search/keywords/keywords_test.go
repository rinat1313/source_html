package keywords

import "testing"

func TestMatchAND(t *testing.T) {
	if !Match("Поставка бумаги", "офисная бумага А4", []string{"бумага", "офис"}) {
		t.Fatal("expected match")
	}
	if Match("Поставка бумаги", "офисная", []string{"бумага", "офис"}) {
		t.Fatal("expected no match")
	}
	if !Match("anything", "x", nil) {
		t.Fatal("empty keywords match all")
	}
}
