package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const (
	DateImportFormat   = "02.01.2006"
	DateImportExpected = "DD.MM.YYYY"
	DateFormat         = "2006-01-02"
	DateExpected       = "YYYY-MM-DD"
)

type MediaEntry struct {
	ID         int64     `json:"id"`
	Date       LocalDate `json:"date"`
	Title      string    `json:"title"`
	IsFinished bool      `json:"isFinished"`
	Type       *string   `json:"type,omitempty"`
	Genre      *string   `json:"genre,omitempty"`
	IsDropped  bool      `json:"isDropped"`
	Comment    *string   `json:"comment,omitempty"`
}

type MediaRating struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	Total    int    `json:"total"`
	Finished int    `json:"finished"`
	Rating   int    `json:"rating"`
}

type TitleStats struct {
	Title       string `json:"title"`
	Last3days   int    `json:"last3Days"`
	Last7days   int    `json:"last7Days"`
	Last30days  int    `json:"last30Days"`
	Last180days int    `json:"last180Days"`
	Total       int    `json:"total"`
}

type PagedResult struct {
	Data       []MediaEntry `json:"data"`
	NextCursor string       `json:"nextCursor,omitempty"`
	HasMore    bool         `json:"hasMore"`
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
	return []byte(fmt.Sprintf(`"%s"`, t.Format(DateImportFormat))), nil
}

func (ld *LocalDate) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	t, err := time.Parse(DateImportFormat, str)
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
	switch v := value.(type) {
	case time.Time:
		*ld = LocalDate(v)
		return nil
	case string:
		t, err := time.Parse(DateImportFormat, v)
		if err != nil {
			return err
		}
		*ld = LocalDate(t)
		return nil
	default:
		return fmt.Errorf("invalid type for LocalDate scan: %T", value)
	}
}

func (ld LocalDate) Value() (driver.Value, error) {
	t := time.Time(ld)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}
