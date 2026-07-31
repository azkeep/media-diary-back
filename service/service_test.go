package service

import (
	"errors"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/repository"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockRepository struct {
	repository.MediaRepository
	importedEntries []model.MediaEntry
	importErr       error
}

func (m *mockRepository) ImportBatch(entries []model.MediaEntry) error {
	m.importedEntries = entries
	return m.importErr
}

func (r CSVRow) toCSVLine() string {
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s",
		r.DateRaw,
		r.TitleRaw,
		r.IsFinishedRaw,
		r.TypeRaw,
		r.GenreRaw,
		r.IsDroppedRaw,
		r.MediaCommentRaw,
	)
}

func buildMockCSV(rows ...CSVRow) *strings.Reader {
	header := "Date;Title;IsFinished;Type;Genre;IsDropped;Comment"
	lines := []string{header}
	for _, row := range rows {
		lines = append(lines, row.toCSVLine())
	}
	return strings.NewReader(strings.Join(lines, "\n"))
}

func TestImportFromCSV(t *testing.T) {
	rowMatrix := CSVRow{
		DateRaw:         "01.01.2026",
		TitleRaw:        "The Matrix",
		IsFinishedRaw:   "true",
		TypeRaw:         "Movie",
		GenreRaw:        "Sci-Fi",
		IsDroppedRaw:    "false",
		MediaCommentRaw: "Classic",
	}

	rowInterstellar := CSVRow{
		DateRaw:         "02.01.2026",
		TitleRaw:        "Interstellar",
		IsFinishedRaw:   "true",
		TypeRaw:         "Movie",
		GenreRaw:        "Sci-Fi",
		IsDroppedRaw:    "false",
		MediaCommentRaw: "Masterpiece",
	}

	tests := []struct {
		name           string
		reader         io.Reader
		repoErr        error
		expectedLen    int
		expectedTitles []string
		wantErr        bool
	}{
		{
			name:           "should import batch entries when CSV payload is valid",
			reader:         buildMockCSV(rowMatrix, rowInterstellar),
			repoErr:        nil,
			expectedLen:    2,
			expectedTitles: []string{rowMatrix.TitleRaw, rowInterstellar.TitleRaw},
			wantErr:        false,
		},
		{
			name:        "should return error when CSV payload is empty",
			reader:      strings.NewReader(""),
			expectedLen: 0,
			wantErr:     true,
		},
		{
			name:        "should return error when repository fails batch import",
			reader:      buildMockCSV(rowMatrix),
			repoErr:     errors.New("db write failed"),
			expectedLen: 1,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{importErr: tt.repoErr}
			svc := NewMediaService(mockRepo)

			err := svc.ImportFromCSV(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ImportFromCSV() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if len(mockRepo.importedEntries) != tt.expectedLen {
					t.Fatalf("expected %d imported entries, got %d", tt.expectedLen, len(mockRepo.importedEntries))
				}
				for i, expectedTitle := range tt.expectedTitles {
					if mockRepo.importedEntries[i].Title != expectedTitle {
						t.Errorf("entry index %d: expected title %q, got %q", i, expectedTitle, mockRepo.importedEntries[i].Title)
					}
				}
			}
		})
	}
}

func TestMapScore(t *testing.T) {
	tests := []struct {
		name     string
		rawScore int
		expected int
	}{
		{"should map to 20 when score is at or above top boundary", 25, 20},
		{"should map to 20 when score is exactly 20", 20, 20},
		{"should map to 10 when score is in tier 10", 15, 10},
		{"should map to 5 when score is in tier 5", 8, 5},
		{"should map to 2 when score is in tier 2", 4, 2},
		{"should map to 1 when score is in tier 1", 2, 1},
		{"should map to 0 when score is default zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapScore(tt.rawScore)
			if got != tt.expected {
				t.Errorf("mapScore(%d) = %d; want %d", tt.rawScore, got, tt.expected)
			}
		})
	}
}

func TestCSVRow_ToDomainModel(t *testing.T) {
	validRow := CSVRow{
		DateRaw:         "15.03.2026",
		TitleRaw:        "Interstellar",
		IsFinishedRaw:   "true",
		TypeRaw:         "Movie",
		GenreRaw:        "Sci-Fi",
		IsDroppedRaw:    "false",
		MediaCommentRaw: "great Nolan space existential",
	}

	tests := []struct {
		name    string
		row     CSVRow
		wantErr bool
	}{
		{
			name:    "should convert successfully when all fields are valid",
			row:     validRow,
			wantErr: false,
		},
		{
			name: "should return error when title is empty",
			row: CSVRow{
				DateRaw:  "15.03.2026",
				TitleRaw: "",
			},
			wantErr: true,
		},
		{
			name: "should return error when date format is invalid",
			row: CSVRow{
				TitleRaw: "Test",
				DateRaw:  "2026-03-15",
			},
			wantErr: true,
		},
		{
			name: "should return error when boolean is invalid",
			row: CSVRow{
				DateRaw:       "15.03.2026",
				TitleRaw:      "Test",
				IsFinishedRaw: "invalid_bool",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.row.ToDomainModel()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ToDomainModel() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got.Title != tt.row.TitleRaw {
					t.Errorf("got title %v, expected %v", got.Title, tt.row.TitleRaw)
				}

				isFinishedExpected, _ := strconv.ParseBool(tt.row.IsFinishedRaw)
				if got.IsFinished != isFinishedExpected {
					t.Errorf("got isFinished %v, expected %v", got.IsFinished, isFinishedExpected)
				}

				isDroppedExpected, _ := strconv.ParseBool(tt.row.IsDroppedRaw)
				if got.IsDropped != isDroppedExpected {
					t.Errorf("got isDropped %v, expected %v", got.IsDropped, isDroppedExpected)
				}

				expectedDate, _ := time.Parse(DateFormat, tt.row.DateRaw)
				if got.Date.Time() != expectedDate {
					t.Errorf("got date %v, expected %v", got.Date.Time(), expectedDate)
				}
			}
		})
	}
}
