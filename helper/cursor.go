package helper

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"api-students/app/model"
)

var ErrInvalidCursor = &cursorError{"invalid cursor"}

type cursorError struct {
	message string
}

func (e *cursorError) Error() string {
	return e.message
}

func EncodeCursor(createdAt time.Time, id string) string {
	raw := strconv.FormatInt(createdAt.UnixNano(), 10) + ":" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return model.Cursor{}, ErrInvalidCursor
	}

	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}

	return model.Cursor{
		CreatedAt: time.Unix(0, nanos).UTC(),
		ID:        parts[1],
	}, nil
}
