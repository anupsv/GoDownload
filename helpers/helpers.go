package helpers

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

// GetFileNameFromURL Helper function to extract file name from URL
func GetFileNameFromURL(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}

// IsValidURL Helper function to validate URLs
func IsValidURL(testURL string) bool {
	parsedURL, err := url.ParseRequestURI(testURL)
	if err != nil {
		return false
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}

func ValidateDirectory(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("Directory %s does not exist", dir)
	}

	// Check for write permissions by attempting to create a temporary file
	tempFile, err := ioutil.TempFile(dir, "write_check_")
	if err != nil {
		return fmt.Errorf("No write permissions to directory %s", dir)
	}
	tempFile.Close()
	os.Remove(tempFile.Name()) // Cleanup the temporary file

	return nil
}
