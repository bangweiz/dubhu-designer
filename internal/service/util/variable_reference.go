package util

import (
	"regexp"
	"strings"
)

var variableRefRegex = regexp.MustCompile(`\{\{var:([^{}]*)\}\}`)

// ExtractVariableIDs returns unique references in order of first appearance.
// Empty IDs are retained so validation rejects empty variable references.
func ExtractVariableIDs(content string) []string {
	ids := []string{}
	seen := make(map[string]bool)
	for _, match := range variableRefRegex.FindAllStringSubmatch(content, -1) {
		id := strings.TrimSpace(match[1])
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}

	return ids
}
