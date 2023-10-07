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

//func TestSegmentMerging(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockHttpClient := NewMockHttpClient(ctrl)
//	downloader := &Downloader{Client: mockHttpClient}
//
//	// Mock the HEAD request to get file size
//	mockHttpClient.EXPECT().Head("http://example.com").Return(&http.Response{
//		StatusCode:    http.StatusOK,
//		ContentLength: 1000, // example content length
//		Body:          http.NoBody,
//	}, nil)
//
//	// Mock the segment downloads
//	for i := 0; i < 4; i++ { // assuming 4 segments for simplicity
//		req, _ := http.NewRequest("GET", "http://example.com", nil)
//		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i*250, (i+1)*250-1))
//		mockHttpClient.EXPECT().Do(gomock.Any(), req).Return(&http.Response{
//			StatusCode: http.StatusOK,
//			Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("segment-%d-data/", i+1))), // unique dummy segment data
//		}, nil)
//	}
//
//	// Create dummy segment files
//	for i := 0; i < 4; i++ {
//		ioutil.WriteFile(fmt.Sprintf("/tmp/file.part%d", i+1), []byte(fmt.Sprintf("segment-%d-data/", i+1)), 0644)
//		defer os.Remove(fmt.Sprintf("/tmp/file.part%d", i+1)) // schedule cleanup after test
//	}
//
//	err := downloader.DownloadFileInSegments("http://example.com", "/tmp/file", 4)
//	if err != nil {
//		t.Errorf("Expected no error, but got: %v", err)
//	}
//
//	// Check if the merged file contains all segment data
//	mergedData, _ := ioutil.ReadFile("/tmp/file")
//	expectedData := "segment-1-data/segment-2-data/segment-3-data/segment-4-data/"
//	if string(mergedData) != expectedData {
//		t.Errorf("Segment merging failed. Expected %s but got %s", expectedData, string(mergedData))
//	}
//
//	// Cleanup the merged file
//	defer os.Remove("/tmp/file")
//}
