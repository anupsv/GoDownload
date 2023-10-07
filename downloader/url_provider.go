package downloader

import (
	"GoDownload/helpers"
	"fmt"
)

// URLProvider is an interface to provide URLs.
type URLProvider interface {
	GetURLs() []string
}

// FileURLProvider provides URLs from a file.
type FileURLProvider struct {
	Filename string
}

// TODO: Implement GetURLs for FileURLProvider

type StaticURLProvider struct {
	URLs []string
}

func (s *StaticURLProvider) GetURLs() []string {
	var validURLs []string
	for _, u := range s.URLs {
		if helpers.IsValidURL(u) {
			validURLs = append(validURLs, u)
		} else {
			fmt.Printf("Warning: Skipping invalid URL: %s\n", u)
		}
	}
	return validURLs
}
