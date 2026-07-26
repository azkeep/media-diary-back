package service

import (
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
	"github.com/azkeep/MediaDiary/backend-go/repository"
)

type MediaService interface {
	GetAllMedia() ([]model.MediaSelected, error)
	GetMediaByDate(date model.LocalDate) ([]model.MediaSelected, error)
	GetMediaLaterThan(date model.LocalDate) ([]model.MediaSelected, error)
	Save(media *model.MediaSelected) error
	Update(media *model.MediaSelected) error
	Delete(id int64) error
	GetTitleStats(title string) (*model.StatsResponse, bool, error)
	GetAllTitlesByRating(months int) ([]model.MediaRating, error)
}

type mediaService struct {
	repo repository.MediaRepository
}

const (
	FinishedRatio = 9
)

func NewMediaService(repo repository.MediaRepository) MediaService {
	return &mediaService{repo: repo}
}

func (s *mediaService) GetAllMedia() ([]model.MediaSelected, error) {
	return s.repo.FindAllByOrderByDateDesc()
}

func (s *mediaService) GetMediaByDate(date model.LocalDate) ([]model.MediaSelected, error) {
	return s.repo.FindByDate(date)
}

func (s *mediaService) GetMediaLaterThan(date model.LocalDate) ([]model.MediaSelected, error) {
	return s.repo.FindAllByDateGreaterThanEqualOrderByDateDesc(date)
}

func (s *mediaService) Save(media *model.MediaSelected) error {
	media.ID = 0 // Ensure ID is generated
	return s.repo.Save(media)
}

func (s *mediaService) Update(media *model.MediaSelected) error {
	exists, err := s.repo.ExistsByID(media.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("media entry does not exist with id: %d", media.ID)
	}
	return s.repo.Save(media)
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
