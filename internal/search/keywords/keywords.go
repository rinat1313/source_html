package keywords

import (
	"strings"

	"github.com/rinat1313/source_html/internal/apitypes"
)

func Match(title, description string, keywords []string) bool {
	words := apitypes.NormalizeKeywords(keywords)
	if len(words) == 0 {
		return true
	}
	hay := strings.ToLower(title + " " + description)
	for _, w := range words {
		if !strings.Contains(hay, w) {
			return false
		}
	}
	return true
}
