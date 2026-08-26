package mcp

import (
	"encoding/base64"
	"fmt"
	"strconv"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// defaultPageSize is the number of items returned per page by paginated
// list methods when no cursor is supplied.
const defaultPageSize = 50

// encodeCursor encodes an opaque pagination cursor for the given offset.
func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

// decodeCursor decodes an opaque pagination cursor into its offset.
func decodeCursor(cursor string) (int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, naeoserr.New(naeoserr.ErrValidation, "invalid params: 'cursor' is not a valid opaque cursor")
	}
	offset, err := strconv.Atoi(string(raw))
	if err != nil || offset < 0 {
		return 0, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid params: 'cursor' %q does not resolve to a valid offset", cursor))
	}
	return offset, nil
}

// paginate returns the page of items starting at the offset encoded in
// cursor, plus the nextCursor to pass for the following page. An empty
// nextCursor means no further pages remain.
func paginate[T any](items []T, cursor string) ([]T, string, error) {
	offset := 0
	if cursor != "" {
		var err error
		if offset, err = decodeCursor(cursor); err != nil {
			return nil, "", err
		}
	}
	if offset > len(items) {
		offset = len(items)
	}
	end := offset + defaultPageSize
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = encodeCursor(end)
	}
	return items[offset:end], next, nil
}
