package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/repository"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type MediaService interface {
	GetAllMedia() ([]model.MediaEntry, error)
	GetMediaByDate(date model.LocalDate) ([]model.MediaEntry, error)
	GetMediaLaterThan(date model.LocalDate) ([]model.MediaEntry, error)
	//Save(media *model.MediaEntry) error
	SaveBatch(entries []model.MediaEntry) error
	Update(media *model.MediaEntry) error
	Delete(id int64) error
	GetTitleStats(title string) (*model.StatsResponse, bool, error)
	GetAllTitlesByRating(months int) ([]model.MediaRating, error)
	ImportFromCSV(r io.Reader) error
	ExportRatingsToCSV(w io.Writer, months int) error
	SearchEntries(searchTerm string) ([]model.MediaEntry, error)
}

type CSVRow struct {
	DateRaw         string
	TitleRaw        string
	IsFinishedRaw   string
	TypeRaw         string
	GenreRaw        string
	IsDroppedRaw    string
	MediaCommentRaw string
}

type mediaService struct {
	repo repository.MediaRepository
}

const (
	FinishedRatio               = 9
	ColumnsExpected             = 7
	CSVComma                    = ';'
	DateFormat                  = "02.01.2006"
	DateExpected                = "DD.MM.YYYY"
	ExportDirPerm   os.FileMode = 0755
	ExportDir                   = "export"
)

func NewMediaService(repo repository.MediaRepository) MediaService {
	return &mediaService{repo: repo}
}

func (s *mediaService) GetAllMedia() ([]model.MediaEntry, error) {
	return s.repo.FindAllByOrderByDateDesc()
}

func (s *mediaService) GetMediaByDate(date model.LocalDate) ([]model.MediaEntry, error) {
	return s.repo.FindByDate(date)
}

func (s *mediaService) GetMediaLaterThan(date model.LocalDate) ([]model.MediaEntry, error) {
	return s.repo.FindAllByDateGreaterThanEqualOrderByDateDesc(date)
}

func (s *mediaService) SaveBatch(entries []model.MediaEntry) error {
	if len(entries) == 0 {
		return errors.New("empty media entries")
	}

	for i := range entries {
		entries[i].ID = 0
		if strings.TrimSpace(entries[i].Title) == "" {
			return fmt.Errorf("entry at index %d has empty title", i)
		}
	}

	return s.repo.SaveBatch(entries)
}

func (s *mediaService) Update(media *model.MediaEntry) error {
	//exists, err := s.repo.ExistsByID(media.ID)
	//if err != nil {
	//	return err
	//}
	//if !exists {
	//	return fmt.Errorf("media entry does not exist with id: %d", media.ID)
	//}
	//return s.repo.Save(media)
	return nil
}

func (s *mediaService) Delete(id int64) error {
	return s.repo.DeleteByID(id)
}

func (s *mediaService) GetTitleStats(title string) (*model.StatsResponse, bool, error) {
	return s.repo.GetTitleStats(title)
}

func (s *mediaService) GetAllTitlesByRating(months int) ([]model.MediaRating, error) {
	ratings, err := s.repo.FindAllUniqueTitlesByRating(months)

	if err != nil || len(ratings) == 0 {
		return ratings, err
	}

	calculateScores(ratings)

	return ratings, nil
}

func (s *mediaService) ImportFromCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.Comma = CSVComma
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("CSV file is empty")
		}
		return fmt.Errorf("Failed to read CSV header: %w", err)
	}

	columnsActual := len(header)
	if columnsActual < ColumnsExpected {
		return fmt.Errorf("Invalid CSV format: expected %d columns, got %d", ColumnsExpected, columnsActual)
	}

	var parsedEntries []model.MediaEntry
	rowNum := 0

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("CSV parse error at row %d: %w", rowNum+1, err)
		}
		rowNum++

		rawRow := newCSVRow(record)

		entry, err := rawRow.ToDomainModel()

		if err != nil {
			return fmt.Errorf("CSV parse error at row %d: %w", rowNum, err)
		}

		parsedEntries = append(parsedEntries, *entry)
	}

	if len(parsedEntries) == 0 {
		return errors.New("CSV file contains no valid data rows")
	}

	return s.repo.ImportBatch(parsedEntries)
}

func (s *mediaService) ExportRatingsToCSV(w io.Writer, months int) error {
	ratings, err := s.GetAllTitlesByRating(months)
	if err != nil {
		return fmt.Errorf("failed to fetch ratings: %w", err)
	}

	if err := os.MkdirAll(ExportDir, ExportDirPerm); err != nil {
		return fmt.Errorf("failed to create `%s` directory: %w", ExportDir, err)
	}

	exportFileName := fmt.Sprintf("%s-ratings-last-%d-months.csv", time.Now().Format("20060102150405"), months)
	exportFilePath := filepath.Join(ExportDir, exportFileName)

	file, err := os.Create(exportFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local export file %s: %w", exportFilePath, err)
	}
	defer file.Close()

	mw := io.MultiWriter(w, file)
	writer := csv.NewWriter(mw)
	writer.Comma = CSVComma

	header := []string{"Title", "Type", "Total", "Finished", "Rating"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	if err := writer.Write([]string{}); err != nil {
		return fmt.Errorf("failed to write empty line: %w", err)
	}

	for _, r := range ratings {
		row := []string{
			r.Title,
			r.Type,
			strconv.Itoa(r.Total),
			strconv.Itoa(r.Finished),
			strconv.Itoa(r.Rating),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return writer.Error()
}

func (s *mediaService) SearchEntries(searchTerm string) ([]model.MediaEntry, error) {
	searchTerm = strings.TrimSpace(searchTerm)
	return s.repo.SearchEntries(searchTerm)
}

func calculateScores(ratings []model.MediaRating) {
	for r := range ratings {
		finished := ratings[r].Finished
		total := ratings[r].Total

		rawScore := FinishedRatio*finished + total

		ratings[r].Rating = mapScore(rawScore / 2)
	}
}

func mapScore(raw int) int {
	switch {
	case raw >= 20:
		return 20
	case raw > 10:
		return 10
	case raw > 5:
		return 5
	case raw > 2:
		return 2
	case raw > 1:
		return 1
	default:
		return 0
	}
}

func newCSVRow(record []string) CSVRow {
	row := CSVRow{}
	if len(record) > 0 {
		row.DateRaw = strings.TrimSpace(record[0])
	}
	if len(record) > 1 {
		row.TitleRaw = strings.TrimSpace(record[1])
	}
	if len(record) > 2 {
		row.IsFinishedRaw = strings.TrimSpace(record[2])
	}
	if len(record) > 3 {
		row.TypeRaw = strings.TrimSpace(record[3])
	}
	if len(record) > 4 {
		row.GenreRaw = strings.TrimSpace(record[4])
	}
	if len(record) > 5 {
		row.IsDroppedRaw = strings.TrimSpace(record[5])
	}
	if len(record) > 6 {
		row.MediaCommentRaw = strings.TrimSpace(record[6])
	}
	return row
}

func (r CSVRow) ToDomainModel() (*model.MediaEntry, error) {
	if r.TitleRaw == "" {
		return nil, errors.New("Title cannot be empty")
	}

	dateStr := strings.TrimSpace(r.DateRaw)
	parsedTime, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return nil, fmt.Errorf("Invalid date: %s (expected %s)", r.DateRaw, DateExpected)
	}

	var isFinished bool
	if r.IsFinishedRaw != "" {
		isFinished, err = strconv.ParseBool(strings.ToLower(r.IsFinishedRaw))
		if err != nil {
			return nil, fmt.Errorf("Invalid boolean fo isFinished: '%s'", r.IsFinishedRaw)
		}
	}

	var typePtr *string
	if r.TypeRaw != "" {
		typePtr = &r.TypeRaw
	}

	var genrePtr *string
	if r.GenreRaw != "" {
		genrePtr = &r.GenreRaw
	}

	var isDropped bool
	if r.IsDroppedRaw != "" {
		isDropped, err = strconv.ParseBool(strings.ToLower(r.IsDroppedRaw))
		if err != nil {
			return nil, fmt.Errorf("Invalid boolean fo isDropped: '%s'", r.IsDroppedRaw)
		}
	}

	return &model.MediaEntry{
		Title:      r.TitleRaw,
		Date:       model.LocalDate(parsedTime),
		IsFinished: isFinished,
		Type:       typePtr,
		Genre:      genrePtr,
		IsDropped:  isDropped,
		Comment:    r.MediaCommentRaw,
	}, nil
}
