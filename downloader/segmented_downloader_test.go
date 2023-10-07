package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSuccessfulSegmentDownloads(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)
	downloader := &Downloader{Client: mockHttpClient}

	// Mock the HEAD request to get file size
	mockHttpClient.EXPECT().Head("http://example.com").Return(&http.Response{
		StatusCode:    http.StatusOK,
		ContentLength: 1000, // example content length
		Body:          http.NoBody,
	}, nil)

	// Mock the segment downloads
	for i := 0; i < 4; i++ { // assuming 4 segments for simplicity
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i*250, (i+1)*250-1))
		mockHttpClient.EXPECT().Do(gomock.Any(), req).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("segment data")), // dummy segment data
		}, nil)
	}

	// Create dummy segment files
	for i := 0; i < 4; i++ {
		ioutil.WriteFile(fmt.Sprintf("/tmp/file.part%d", i+1), []byte("segment data"), 0644)
		defer os.Remove(fmt.Sprintf("/tmp/file.part%d", i+1)) // schedule cleanup after test
	}

	err := downloader.DownloadFileInSegments("http://example.com", "/tmp/file", 4)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

func TestFailedSegmentDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)
	downloader := &Downloader{Client: mockHttpClient}

	// Mock the HEAD request to get file size
	mockHttpClient.EXPECT().Head("http://example.com").Return(&http.Response{
		StatusCode:    http.StatusOK,
		ContentLength: 1000, // example content length
		Body:          http.NoBody,
	}, nil)

	// Mock the segment downloads
	for i := 0; i < 3; i++ { // first 3 segments are successful
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i*250, (i+1)*250-1))
		mockHttpClient.EXPECT().Do(gomock.Any(), req).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("segment data")), // dummy segment data
		}, nil)
	}

	// The fourth segment download fails
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Range", "bytes=750-999")
	mockHttpClient.EXPECT().Do(gomock.Any(), req).Return(nil, errors.New("download error"))

	// Create dummy segment files for the first 3 segments
	for i := 0; i < 3; i++ {
		ioutil.WriteFile(fmt.Sprintf("/tmp/file.part%d", i+1), []byte("segment data"), 0644)
		defer os.Remove(fmt.Sprintf("/tmp/file.part%d", i+1)) // schedule cleanup after test
	}

	err := downloader.DownloadFileInSegments("http://example.com", "/tmp/file", 4)
	if err == nil {
		t.Errorf("Expected an error due to failed segment download, but got none")
	}
}

func TestHEADRequestFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)
	downloader := &Downloader{Client: mockHttpClient}

	// Mock the HEAD request to get file size
	mockHttpClient.EXPECT().Head("http://example.com").Return(nil, errors.New("failed to get file size"))

	err := downloader.DownloadFileInSegments("http://example.com", "/tmp/file", 4)
	if err == nil {
		t.Errorf("Expected an error due to failed HEAD request, but got none")
	}
}

func TestMergeSegments(t *testing.T) {
	// Create dummy segment files
	destPath := "/tmp/mergedFile"
	segmentCount := 4
	for i := 1; i <= segmentCount; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s.part%d", destPath, i), []byte(fmt.Sprintf("segment%d data", i)), 0644)
		defer os.Remove(fmt.Sprintf("%s.part%d", destPath, i)) // schedule cleanup after test
	}

	// Call the mergeSegments function
	err := mergeSegments(destPath, segmentCount)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check if the merged file contains all segment data
	mergedData, _ := ioutil.ReadFile(destPath)
	expectedData := "segment1 datasegment2 datasegment3 datasegment4 data"
	if string(mergedData) != expectedData {
		t.Errorf("Segment merging failed. Expected %s but got %s", expectedData, string(mergedData))
	}

	// Cleanup the merged file
	defer os.Remove(destPath)
}

func TestMergeSegments_ErrorOpeningSegment(t *testing.T) {
	destPath := "/tmp/mergedFileError1"
	segmentCount := 4

	// Only create 3 out of 4 segment files to simulate an error
	for i := 1; i < segmentCount; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s.part%d", destPath, i), []byte(fmt.Sprintf("segment%d data", i)), 0644)
		defer os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}

	err := mergeSegments(destPath, segmentCount)
	if err == nil {
		t.Errorf("Expected an error due to missing segment, but got none")
	}
}

func TestMergeSegments_ErrorCopyingSegment(t *testing.T) {
	destPath := "/tmp/mergedFileError2"
	segmentCount := 4

	// Create segment files with restricted permissions to simulate a copy error
	for i := 1; i <= segmentCount; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s.part%d", destPath, i), []byte(fmt.Sprintf("segment%d data", i)), 0000)
		defer os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}

	err := mergeSegments(destPath, segmentCount)
	if err == nil {
		t.Errorf("Expected an error due to copy failure, but got none")
	}
}

func TestMergeSegments_ErrorCreatingMergedFile(t *testing.T) {
	destPath := "/nonexistentpath/mergedFileError3"
	segmentCount := 4

	// Create segment files
	for i := 1; i <= segmentCount; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s.part%d", destPath, i), []byte(fmt.Sprintf("segment%d data", i)), 0644)
		defer os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}

	err := mergeSegments(destPath, segmentCount)
	if err == nil {
		t.Errorf("Expected an error due to failure in creating merged file, but got none")
	}
}
