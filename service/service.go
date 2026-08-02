package service

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/repository"
)

type MediaService interface {
	//GetAllMedia() ([]model.MediaEntry, error)
	GetMediaPaginated(cursor string, limit int) (*model.CursorResponse, error)
	GetMediaByDate(date model.LocalDate) (*model.CursorResponse, error)
	GetMediaForNDays(date model.LocalDate) ([]model.MediaEntry, error)
	GetMediaForNDaysPaginated(date model.LocalDate, cursor string, limit int) (*model.CursorResponse, error)
	SaveBatch(entries []model.MediaEntry) error
	UpdateBatch(entries []model.MediaEntry) ([]model.MediaEntry, error)
	DeleteBatch(ids []int64) error
	GetTitleStats(title string) (*model.StatsResponse, bool, error)
	GetAllTitlesByRating(months int) ([]model.MediaRating, error)
	ImportFromCSV(r io.Reader) error
	ExportRatingsToCSV(w io.Writer, months int) error
	SearchEntriesPaginated(searchTerm string, cursor string, limit int) (*model.CursorResponse, error)
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

type parsedCursor struct {
	LastDate   *model.LocalDate
	LastID     int64
	TotalCount *int
}

const (
	FinishedRatio               = 9
	ColumnsExpected             = 7
	CSVComma                    = ';'
	DateFormat                  = "02.01.2006"
	DateExpected                = "DD.MM.YYYY"
	ExportDirPerm   os.FileMode = 0755
	ExportDir                   = "export"
	PageLimit                   = 50
)

func NewMediaService(repo repository.MediaRepository) MediaService {
	return &mediaService{repo: repo}
}

//func (s *mediaService) GetAllMedia() ([]model.MediaEntry, error) {
//	return s.repo.FindAllByOrderByDateDesc()
//}

func (s *mediaService) GetMediaPaginated(encodedCursor string, limit int) (*model.CursorResponse, error) {
	limit = normalizeLimit(limit)
	pc := decodeCursor(encodedCursor)

	if pc.TotalCount == nil {
		count, err := s.repo.CountAllEntries()
		if err == nil {
			pc.TotalCount = &count
		}
	}

	// Fetch 1 extra record to evaluate `HasMore` efficiently without a COUNT(*)
	entries, err := s.repo.FindAllByCursor(pc.LastDate, pc.LastID, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := validateHasMore(entries, limit)
	entries = truncateEntries(entries, limit)
	nextCursor := resolveNextCursor(entries, hasMore, pc.TotalCount)

	return &model.CursorResponse{
		Data:       entries,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      pc.TotalCount,
	}, nil

}

func (s *mediaService) GetMediaByDate(date model.LocalDate) (*model.CursorResponse, error) {
	entries, err := s.repo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	entries = truncateEntries(entries, len(entries))

	total := len(entries)
	return &model.CursorResponse{
		Data:       entries,
		NextCursor: "",
		HasMore:    false,
		Total:      &total,
	}, nil
}

func (s *mediaService) GetMediaForNDays(date model.LocalDate) ([]model.MediaEntry, error) {
	return s.repo.FindAllByDateGreaterThanEqualOrderByDateDesc(date)
}

func (s *mediaService) GetMediaForNDaysPaginated(date model.LocalDate, encodedCursor string, limit int) (*model.CursorResponse, error) {
	limit = normalizeLimit(limit)
	pc := decodeCursor(encodedCursor)

	if pc.TotalCount == nil {
		count, err := s.repo.CountEntriesForNDays(date)
		if err == nil {
			pc.TotalCount = &count
		}
	}

	entries, err := s.repo.FindAllByDateGreaterThanEqualPaginated(date, pc.LastDate, pc.LastID, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := validateHasMore(entries, limit)
	entries = truncateEntries(entries, limit)
	nextCursor := resolveNextCursor(entries, hasMore, pc.TotalCount)

	return &model.CursorResponse{
		Data:       entries,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      pc.TotalCount,
	}, nil
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

func (s *mediaService) UpdateBatch(entries []model.MediaEntry) ([]model.MediaEntry, error) {
	if len(entries) == 0 {
		return nil, errors.New("empty media entries")
	}

	for i := range entries {
		if entries[i].ID <= 0 {
			return nil, fmt.Errorf("entry at index %d has invalid ID: %d", i, entries[i].ID)
		}
		if strings.TrimSpace(entries[i].Title) == "" {
			return nil, fmt.Errorf("entry at index %d has empty title", i)
		}
	}

	updatedEntries, err := s.repo.UpdateBatch(entries)
	if err != nil {
		return nil, fmt.Errorf("failed to update media entries batch: %w", err)
	}

	return updatedEntries, nil
}

func (s *mediaService) DeleteBatch(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("ids list is empty")
	}

	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("invalid entry id: %d", id)
		}
	}

	return s.repo.DeleteBatch(ids)
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
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	columnsActual := len(header)
	if columnsActual < ColumnsExpected {
		return fmt.Errorf("invalid CSV format: expected %d columns, got %d", ColumnsExpected, columnsActual)
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

func (s *mediaService) SearchEntriesPaginated(searchTerm string, encodedCursor string, limit int) (*model.CursorResponse, error) {
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return &model.CursorResponse{
			Data:    []model.MediaEntry{},
			HasMore: false,
		}, nil
	}

	if limit <= 0 {
		entries, err := s.repo.SearchEntries(searchTerm)
		if err != nil {
			return nil, err
		}

		entries = truncateEntries(entries, len(entries))
		total := len(entries)

		return &model.CursorResponse{
			Data:    entries,
			HasMore: false,
			Total:   &total,
		}, nil
	}

	limit = normalizeLimit(limit)
	pc := decodeCursor(encodedCursor)

	if pc.TotalCount == nil {
		count, err := s.repo.CountSearchEntries(searchTerm)
		if err == nil {
			pc.TotalCount = &count
		}
	}

	entries, err := s.repo.SearchEntriesPaginated(searchTerm, pc.LastDate, pc.LastID, limit+1)
	if err != nil {
		return nil, err
	}

	hasMore := validateHasMore(entries, limit)
	entries = truncateEntries(entries, limit)
	nextCursor := resolveNextCursor(entries, hasMore, pc.TotalCount)

	return &model.CursorResponse{
		Data:       entries,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      pc.TotalCount,
	}, nil
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
		return nil, errors.New("title cannot be empty")
	}

	dateStr := strings.TrimSpace(r.DateRaw)
	parsedTime, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s (expected %s)", r.DateRaw, DateExpected)
	}

	var isFinished bool
	if r.IsFinishedRaw != "" {
		isFinished, err = strconv.ParseBool(strings.ToLower(r.IsFinishedRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid boolean fo isFinished: '%s'", r.IsFinishedRaw)
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

	var commentPtr *string
	if r.MediaCommentRaw != "" {
		commentPtr = &r.MediaCommentRaw
	}

	var isDropped bool
	if r.IsDroppedRaw != "" {
		isDropped, err = strconv.ParseBool(strings.ToLower(r.IsDroppedRaw))
		if err != nil {
			return nil, fmt.Errorf("invalid boolean fo isDropped: '%s'", r.IsDroppedRaw)
		}
	}

	return &model.MediaEntry{
		Title:      r.TitleRaw,
		Date:       model.LocalDate(parsedTime),
		IsFinished: isFinished,
		Type:       typePtr,
		Genre:      genrePtr,
		IsDropped:  isDropped,
		Comment:    commentPtr,
	}, nil
}

func decodeCursor(encoded string) *parsedCursor {
	if encoded == "" {
		return &parsedCursor{}
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return &parsedCursor{}
	}

	parts := strings.Split(string(decoded), "_")
	if len(parts) < 2 {
		return &parsedCursor{}
	}

	res := &parsedCursor{}

	if t, err := time.Parse(DateFormat, parts[0]); err == nil {
		ld := model.LocalDate(t)
		res.LastDate = &ld
	}

	res.LastID, _ = strconv.ParseInt(parts[1], 10, 64)

	if len(parts) >= 3 {
		if total, err := strconv.Atoi(parts[2]); err == nil {
			res.TotalCount = &total
		}
	}

	return res
}

func encodeCursor(lastEntry model.MediaEntry, totalCount *int) string {
	var raw string
	if totalCount != nil {
		raw = fmt.Sprintf("%s_%d_%d", lastEntry.Date.Time().Format(DateFormat), lastEntry.ID, *totalCount)
	} else {
		raw = fmt.Sprintf("%s_%d", lastEntry.Date.Time().Format(DateFormat), lastEntry.ID)
	}
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func validateHasMore(entries []model.MediaEntry, limit int) bool {
	return len(entries) > limit
}

func truncateEntries(entries []model.MediaEntry, limit int) []model.MediaEntry {
	if len(entries) > limit {
		entries = entries[:limit]
	}
	if entries == nil {
		return []model.MediaEntry{}
	}
	return entries
}

func resolveNextCursor(entries []model.MediaEntry, hasMore bool, totalCount *int) string {
	if !hasMore || len(entries) == 0 {
		return ""
	}
	return encodeCursor(entries[len(entries)-1], totalCount)
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > (PageLimit*2) {
		return PageLimit
	}
	return limit
}
