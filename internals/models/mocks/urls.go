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

// ShortenerDataModel implements a mock model for testing shortner data
type ShortenerDataModel struct {
}

// Insert mocks the creation of a new shortener data record
func (m *ShortenerDataModel) Insert(original string, shortened string, clicks int) (int, error) {
	return 1, nil
}

// Get mocks the retrieval of a new shortener data record
func (m *ShortenerDataModel) Get(shortened string) (*models.ShortenerData, error) {
	switch shortened {
	case "http://shorten3d":
		return mockDataModel, nil
	default:
		return nil, errors.New("models: no matching record found")
	}
}

// IncrementClicks mocks incrementing the click cound for a shortener data record
func (m *ShortenerDataModel) IncrementClicks(shortened string) error {
	switch shortened {
	case "http://shorten3d":
		return nil
	default:
		return errors.New("models: no matching record found")
	}
}

// Latest mocks incrementing retrieving all shortener data records
func (m *ShortenerDataModel) Latest() ([]*models.ShortenerData, error) {
	return []*models.ShortenerData{mockDataModel}, nil
}
