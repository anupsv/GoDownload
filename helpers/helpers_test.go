package helpers

import (
	"io/ioutil"
	"os"
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
		result := GetFileNameFromURL(tt.url)
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
		result := IsValidURL(tt.url)
		if result != tt.expected {
			t.Errorf("isValidURL(%q) = %v; want %v", tt.url, result, tt.expected)
		}
	}
}

func TestValidateDirectory(t *testing.T) {
	// 1. Test with a valid directory
	tempDir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = ValidateDirectory(tempDir)
	if err != nil {
		t.Errorf("Expected no error for valid directory, got: %v", err)
	}

	// 2. Test with a directory with no write permissions
	noWriteDir, err := ioutil.TempDir("", "testnowritedir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(noWriteDir)

	os.Chmod(noWriteDir, 0555) // Remove write permissions
	err = ValidateDirectory(noWriteDir)
	if err == nil {
		t.Error("Expected error for directory with no write permissions, got nil")
	}

	// 3. Test with a non-existent directory
	err = ValidateDirectory("/path/to/non/existent/dir")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}
