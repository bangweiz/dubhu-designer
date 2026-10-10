package util

import (
	"regexp"
	"strings"
)

var toolRefRegex = regexp.MustCompile(`\{\{tool:([^{}]+)\}\}`)

// ExtractToolIDs parses the content string and extracts all unique tool IDs referenced via {{tool:tool_id}}.
func ExtractToolIDs(content string) []string {
	matches := toolRefRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{})
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			id := strings.TrimSpace(m[1])
			if id != "" {
				if _, exists := seen[id]; !exists {
					seen[id] = struct{}{}
					ids = append(ids, id)
				}
			}
		}
	}

	return ids
}
