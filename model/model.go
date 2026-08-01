package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type MediaEntry struct {
	ID         int64     `json:"id"`
	Date       LocalDate `json:"date"`
	Title      string    `json:"title"`
	IsFinished bool      `json:"isFinished"`
	Type       *string   `json:"type"`
	Genre      *string   `json:"genre"`
	IsDropped  bool      `json:"isDropped"`
	Comment    *string   `json:"comment"`
}

type MediaRating struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Total    int    `json:"total"`
	Finished int    `json:"finished"`
	Rating   int    `json:"rating"`
}

type StatsResponse struct {
	Title       string `json:"title"`
	Last3days   int    `json:"last_3_days"`
	Last7days   int    `json:"last_7_days"`
	Last30days  int    `json:"last_30_days"`
	Last180days int    `json:"last_180_days"`
	Total       int    `json:"total"`
}

type CursorResponse struct {
	Data       []MediaEntry `json:"data"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
	Total      *int         `json:"total,omitempty"`
}

// LocalDate handles YYYY-MM-DD format for JSON and DB operations
type LocalDate time.Time

func (ld LocalDate) Time() time.Time {
	return time.Time(ld)
}

func (ld LocalDate) MarshalJSON() ([]byte, error) {
	t := time.Time(ld)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, t.Format("2006-01-02"))), nil
}

func (ld *LocalDate) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	*ld = LocalDate(t)
	return nil
}

func (ld *LocalDate) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("invalid type for LocalDate scan: %T", value)
	}
	*ld = LocalDate(t)
	return nil
}

func (ld LocalDate) Value() (driver.Value, error) {
	t := time.Time(ld)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}
