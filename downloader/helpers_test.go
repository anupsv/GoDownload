package downloader

import (
	"testing"
)

func TestGetFileNameFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/file.txt", "file.txt"},
		{"https://example.com/path/to/file.jpg", "file.jpg"},
		{"https://example.com/", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := getFileNameFromURL(tt.url)
		if result != tt.expected {
			t.Errorf("getFileNameFromURL(%q) = %q; want %q", tt.url, result, tt.expected)
		}
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/file.txt", true},
		{"http://example.com", true},
		{"ftp://example.com", true},
		{"example.com", false},
		{"https:/example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidURL(tt.url)
		if result != tt.expected {
			t.Errorf("isValidURL(%q) = %v; want %v", tt.url, result, tt.expected)
		}
	}
}
