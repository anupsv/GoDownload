package downloader

import (
	"bytes"
	"context"
	"errors"
	"github.com/cheggaaa/pb/v3"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

var setupOnce sync.Once
var ctx context.Context
var sugar *zap.SugaredLogger

func setup() {
	// This code will be run once, regardless of how many tests call the setup function.
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar = logger.Sugar()
	ctx = context.WithValue(context.Background(), "sugar", sugar)
}

func TestDownloader_DownloadFile_Success(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockHttpClient(ctrl)
	mockClient.EXPECT().Get("https://www.example.com").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("Hello World")),
	}, nil).Times(1)

	downloader := New(mockClient)
	bar := pb.New64(100)

	err := downloader.DownloadFile("https://www.example.com", "testfile.txt", bar, ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if "testfile.txt" exists and contains "Hello World"
	content, _ := ioutil.ReadFile("testfile.txt")
	if string(content) != "Hello World" {
		t.Fatalf("Expected file content to be 'Hello World', got %s", content)
	}

	// Cleanup after test
	os.Remove("testfile.txt")
}

func TestDownloadFile_FileAlreadyExists(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockHttpClient(ctrl)
	downloader := New(mockClient)

	// Mock HTTP client to simulate a successful file download
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader("file content")),
	}
	mockClient.EXPECT().Get("https://example.com/file.txt").Return(mockResponse, nil).Times(0) // Expect no call
	bar := pb.New64(100)

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Create a dummy file in the temporary directory to simulate a pre-existing file
	destPath := filepath.Join(tempDir, "file.txt")
	ioutil.WriteFile(destPath, []byte("existing content"), 0666)

	// Create a downloader instance with the mock HTTP client

	// Call DownloadFile
	err = downloader.DownloadFile("https://example.com/file.txt", destPath, bar, ctx)
	if err != nil {
		t.Fatalf("Error in DownloadFile: %v", err)
	}

	// Verify that the file content hasn't changed
	content, _ := ioutil.ReadFile(destPath)
	if string(content) != "existing content" {
		t.Fatalf("Expected file content to remain unchanged, got: %s", content)
	}
}

func TestDownloader_DownloadFile_HttpError(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockHttpClient(ctrl)
	mockClient.EXPECT().Get("https://www.example.com/notfound").Return(&http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Body:       io.NopCloser(bytes.NewBufferString("Not Found")),
	}, nil).Times(1)

	downloader := New(mockClient)
	bar := pb.New64(100)
	err := downloader.DownloadFile("https://www.example.com/notfound", "testfile.txt", bar, ctx)
	if err == nil || err.Error() != "failed to download file: 404 Not Found" {
		t.Fatalf("Expected error 'failed to download file: 404 Not Found', got %v", err)
	}
}

func TestDownloader_DownloadFile_ClientError(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockHttpClient(ctrl)
	mockClient.EXPECT().Get("https://www.example.com").Return(nil, errors.New("client error")).Times(1)

	downloader := New(mockClient)
	bar := pb.New64(100)
	err := downloader.DownloadFile("https://www.example.com", "testfile.txt", bar, ctx)
	if err == nil || err.Error() != "client error" {
		t.Fatalf("Expected error 'client error', got %v", err)
	}
}

func TestDownloadFiles(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)

	// Mock the HEAD request to return a Content-Length of 12 (length of "file content")
	mockHeadResponse := &http.Response{
		StatusCode:    http.StatusOK,
		ContentLength: int64(len("file content")),
		Body:          ioutil.NopCloser(bytes.NewReader([]byte{})),
	}
	mockHttpClient.EXPECT().Head("https://example.com/file.txt").Return(mockHeadResponse, nil).Times(1)

	// Mock the GET request to simulate a file download
	mockGetResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("file content"))),
	}
	mockHttpClient.EXPECT().Get("https://example.com/file.txt").Return(mockGetResponse, nil).Times(1)

	// Create a downloader instance with the mock HTTP client
	dl := New(mockHttpClient)

	// Create a StaticURLProvider with one URL
	provider := &StaticURLProvider{URLs: []string{"https://example.com/file.txt"}}

	// Define a temporary directory for downloads
	tempDir, err := ioutil.TempDir("", "testDownload")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Call DownloadFiles
	bars := dl.DownloadFiles(provider, tempDir, 1, ctx)

	// Assertion 1: Check if the file was saved
	destPath := filepath.Join(tempDir, "file.txt")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatalf("File was not saved to %s", destPath)
	}

	// Assertion 2: Check the content of the saved file
	content, err := ioutil.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if string(content) != "file content" {
		t.Fatalf("Expected file content to be 'file content', got: %s", content)
	}

	// Assertion 3: Check if progress bar reached 100%
	for _, bar := range bars {
		if bar.Current() != bar.Total() {
			t.Fatalf("Progress bar did not reach 100%%. Current: %d, Total: %d", bar.Current(), bar.Total())
		}
	}
}

func TestDownloadFiles_HeadError(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)

	// Mock the HEAD request to return an error
	mockHttpClient.EXPECT().Head("https://example.com/file.txt").Return(nil, errors.New("HEAD request failed")).Times(1)

	// Create a downloader instance with the mock HTTP client
	dl := New(mockHttpClient)

	// Create a StaticURLProvider with one URL
	provider := &StaticURLProvider{URLs: []string{"https://example.com/file.txt"}}

	// Define a temporary directory for downloads
	tempDir, err := ioutil.TempDir("", "testDownload")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Call DownloadFiles
	bars := dl.DownloadFiles(provider, tempDir, 1, ctx)

	// Assertion: Check if the file was not saved
	destPath := filepath.Join(tempDir, "file.txt")
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Fatalf("File should not have been saved to %s", destPath)
	}

	// Assertion: Check if progress bar is nil
	for _, bar := range bars {
		if bar != nil {
			t.Fatalf("Expected progress bar to be nil due to HEAD request error")
		}
	}
}

func TestDownloadFile_ErrorCreatingFile(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)

	// Mock the GET request to return a successful response
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("file content"))),
	}
	mockHttpClient.EXPECT().Get("https://example.com/file.txt").Return(mockResponse, nil).Times(1)

	// Create a downloader instance with the mock HTTP client
	dl := New(mockHttpClient)

	// Use an invalid path to simulate an error when creating the file
	invalidPath := "/invalid_path/file.txt"

	// Call DownloadFile
	err := dl.DownloadFile("https://example.com/file.txt", invalidPath, nil, ctx)

	// Assertion: Check if there's an error when trying to create the file
	if err == nil || !strings.Contains(err.Error(), "open /invalid_path/file.txt: no such file or directory") {
		t.Fatalf("Expected error when creating file, got: %v", err)
	}
}
