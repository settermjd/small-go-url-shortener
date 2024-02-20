package mocks

import (
	"errors"
	"gourlshortener/internals/models"
)

var mockDataModel = &models.ShortenerData{
	OriginalURL:  "https://osnews.com",
	ShortenedURL: "http://shorten3d",
	Clicks:       2120,
}

type ShortenerDataModel struct {
}

func (m *ShortenerDataModel) Insert(original string, shortened string, clicks int) (int, error) {
	return 1, nil
}

func (m *ShortenerDataModel) Get(shortened string) (*models.ShortenerData, error) {
	switch shortened {
	case "http://shorten3d":
		return mockDataModel, nil
	default:
		return nil, errors.New("models: no matching record found")
	}
}

func (m *ShortenerDataModel) IncrementClicks(shortened string) error {
	switch shortened {
	case "http://shorten3d":
		return nil
	default:
		return errors.New("models: no matching record found")
	}
}

func (m *ShortenerDataModel) Latest() ([]*models.ShortenerData, error) {
	return []*models.ShortenerData{mockDataModel}, nil
}
