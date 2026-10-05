package etag

import (
	"errors"
	"strconv"
)

var ErrInvalid = errors.New("invalid ETag")

// Format returns a strong ETag containing the resource version.
func Format(version int) string {
	return `"` + strconv.Itoa(version) + `"`
}

// Parse extracts a positive version from a single strong ETag.
func Parse(value string) (int, error) {
	if len(value) < 3 || value[0] != '"' || value[len(value)-1] != '"' {
		return 0, ErrInvalid
	}

	version, err := strconv.Atoi(value[1 : len(value)-1])
	if err != nil || version <= 0 {
		return 0, ErrInvalid
	}
	return version, nil
}
