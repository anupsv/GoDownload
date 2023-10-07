package main

import (
	"GoDownload/downloader"
	"context"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

//func TestRunDownloader_HelpFlag(t *testing.T) {
//	err := RunDownloader(true, 1, "./", nil)
//	if err != nil {
//		t.Fatalf("Expected no error with helpFlag, got %v", err)
//	}
//}
//
//func TestRunDownloader_NoURLs(t *testing.T) {
//	err := RunDownloader(false, 1, "./", []string{})
//	if err == nil || err.Error() != "Please provide URLs to download using the -url flag." {
//		t.Fatalf("Expected error for no URLs, got %v", err)
//	}
//}

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

func TestRunDownloader_ValidURLs(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDownloader := downloader.NewMockDownloaderInterface(ctrl)
	mockDownloader.EXPECT().DownloadFiles(gomock.Any(), "./", 1, gomock.Any()).Times(1)

	mockFactory := downloader.NewMockDownloaderFactory(ctrl)
	mockFactory.EXPECT().NewDownloader(gomock.Any()).Return(mockDownloader).Times(1)

	err := RunDownloader(false, 1, "./", []string{"https://example.com/file1.txt"}, mockFactory, ctx)
	if err != nil {
		t.Fatalf("Expected no error with valid URL, got %v", err)
	}
}

func TestRunDownloader_MultipleValidURLs(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDownloader := downloader.NewMockDownloaderInterface(ctrl)
	mockDownloader.EXPECT().DownloadFiles(gomock.Any(), "./", 2, gomock.Any()).Times(1)

	mockFactory := downloader.NewMockDownloaderFactory(ctrl)
	mockFactory.EXPECT().NewDownloader(gomock.Any()).Return(mockDownloader).Times(1)

	urls := []string{"https://example.com/file1.txt", "https://example.com/file2.txt"}
	err := RunDownloader(false, 2, "./", urls, mockFactory, ctx)
	if err != nil {
		t.Fatalf("Expected no error with multiple valid URLs, got %v", err)
	}
}

func TestRunDownloader_MoreThreadsThanURLs(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDownloader := downloader.NewMockDownloaderInterface(ctrl)
	mockDownloader.EXPECT().DownloadFiles(gomock.Any(), "./", 1, gomock.Any()).Times(1)

	mockFactory := downloader.NewMockDownloaderFactory(ctrl)
	mockFactory.EXPECT().NewDownloader(gomock.Any()).Return(mockDownloader).Times(1)

	urls := []string{"https://example.com/file1.txt"}
	err := RunDownloader(false, 5, "./", urls, mockFactory, ctx)
	if err != nil {
		t.Fatalf("Expected no error with more threads than URLs, got %v", err)
	}
}

func TestRunDownloader_InvalidDirectory(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := downloader.NewMockDownloaderFactory(ctrl)

	urls := []string{"https://example.com/file1.txt"}
	err := RunDownloader(false, 1, "/invalid_directory/", urls, mockFactory, ctx)
	if err == nil {
		t.Fatalf("Expected an error with an invalid directory, got nil")
	}
}

func TestRunDownloader_NoWritePermissions(t *testing.T) {
	setupOnce.Do(setup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup after test

	// Set read-only permissions
	err = os.Chmod(tempDir, 0444) // Read-only permissions
	if err != nil {
		t.Fatalf("Failed to set read-only permissions: %v", err)
	}

	mockFactory := downloader.NewMockDownloaderFactory(ctrl)

	urls := []string{"https://example.com/file1.txt"}
	err = RunDownloader(false, 1, tempDir, urls, mockFactory, ctx)
	if err == nil {
		t.Fatalf("Expected an error due to no write permissions, got nil")
	}
}
