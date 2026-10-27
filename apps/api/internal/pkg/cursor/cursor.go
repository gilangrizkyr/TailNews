package cursor

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"news-portal/apps/api/internal/domain/article"
)

var ErrInvalidCursor = errors.New("invalid cursor format")

// Encode converts article published timestamp and ID into an opaque, URL-safe base64 cursor token.
func Encode(publishedAt time.Time, id uint64) string {
	raw := fmt.Sprintf("%s,%d", publishedAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// Decode extracts published timestamp and ID from an opaque URL-safe base64 cursor token.
// Returns (nil, nil) if cursor token is empty.
func Decode(token string) (*article.Cursor, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil
	}

	decodedBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	parts := strings.SplitN(string(decodedBytes), ",", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidCursor
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		// Fallback to standard RFC3339 if nano fails
		parsedTime, err = time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return nil, ErrInvalidCursor
		}
	}

	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return &article.Cursor{
		PublishedAt: parsedTime.UTC(),
		ID:          id,
	}, nil
}
