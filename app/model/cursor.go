package model

import "time"

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

type CursorQuery struct {
	Search   string
	Active   *bool
	MinGrade *float64
	MaxGrade *float64
	Limit    int
	After    *Cursor
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
