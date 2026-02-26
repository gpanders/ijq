package overlay

import (
	"strconv"
	"strings"
)

func filterIndexes(entries []string, query string) []int {
	if len(entries) == 0 {
		return nil
	}

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		indexes := make([]int, len(entries))
		for i := range entries {
			indexes[i] = i
		}

		return indexes
	}

	needle := strings.ToLower(trimmed)
	indexes := make([]int, 0, len(entries))
	for i, entry := range entries {
		if strings.Contains(strings.ToLower(entry), needle) {
			indexes = append(indexes, i)
		}
	}

	return indexes
}

func formatCount(shown int, total int) string {
	if shown < 0 {
		shown = 0
	}

	if total < 0 {
		total = 0
	}

	return "showing " + strconv.Itoa(shown) + " of " + strconv.Itoa(total) + " entries"
}
