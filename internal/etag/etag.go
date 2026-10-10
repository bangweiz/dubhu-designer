package etag

import (
	"errors"
	"time"
)

var ErrInvalid = errors.New("invalid ETag")

// Format returns a strong ETag containing the persisted updatedAt timestamp.
func Format(updatedAt time.Time) string {
	return `"` + updatedAt.UTC().Truncate(time.Millisecond).Format(time.RFC3339Nano) + `"`
}

// Parse extracts a millisecond-precision timestamp from a single strong ETag.
func Parse(value string) (time.Time, error) {
	if len(value) < 3 || value[0] != '"' || value[len(value)-1] != '"' {
		return time.Time{}, ErrInvalid
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, value[1:len(value)-1])
	if err != nil || updatedAt.IsZero() || !updatedAt.Equal(updatedAt.Truncate(time.Millisecond)) {
		return time.Time{}, ErrInvalid
	}

	return updatedAt.UTC(), nil
}
