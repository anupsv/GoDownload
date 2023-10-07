package downloader

import (
	"net/url"
	"strings"
)

// Helper function to extract file name from URL
func getFileNameFromURL(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}

// Helper function to validate URLs
func isValidURL(testURL string) bool {
	parsedURL, err := url.ParseRequestURI(testURL)
	if err != nil {
		return false
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}
