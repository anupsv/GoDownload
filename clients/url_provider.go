package clients

import (
	"GoDownload/helpers"
	"fmt"
)

// URLProvider is an interface to provide URLs.
type URLProvider interface {
	GetURLs() ([]string, error)
}

// FileURLProvider provides URLs from a file.
type FileURLProvider struct {
	Filename string
}

type StaticURLProvider struct {
	URLs []string
}

func (s *StaticURLProvider) GetURLs() ([]string, error) {
	var validURLs []string
	for _, u := range s.URLs {
		if helpers.IsValidURL(u) {
			validURLs = append(validURLs, u)
		} else {
			return nil, fmt.Errorf("warning: Skipping invalid URL: %s", u)
		}
	}
	return validURLs, nil
}
