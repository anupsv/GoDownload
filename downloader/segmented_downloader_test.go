package downloader

import (
	"GoDownload/clients"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSuccessfulSegmentDownloads(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := clients.NewMockHttpClient(ctrl)
	mockFactory := NewMockSegmentManagerFactory(ctrl)
	mockSegmentManager := NewMockSegmentManager(ctrl)

	url := "http://example.com/file.zip"
	destPath := "/tmp/file.zip"
	segments := 3

	// Mock the HEAD request to get file size
	mockResp := &http.Response{
		StatusCode:    http.StatusOK,
		ContentLength: 900, // Let's assume each segment is 300 bytes
		Body:          http.NoBody,
	}
	mockClient.EXPECT().Head(url).Return(mockResp, nil)

	// Mock the factory to return our mock segment manager
	mockFactory.EXPECT().NewSegmentManager(mockClient, url, destPath).Return(mockSegmentManager)

	// Mock the segment downloads
	for i := 0; i < segments; i++ {
		start := int64(i * 300)
		end := int64((i+1)*300 - 1)
		mockSegmentManager.EXPECT().DownloadSegment(gomock.Any(), mockClient, url, start, end, gomock.Any()).Return(nil)
	}

	// Mock the segment merge
	mockSegmentManager.EXPECT().MergeSegments(destPath, segments).Return(nil)

	downloader := NewSegmentedDownloader(mockClient, mockFactory, url, destPath)
	err := downloader.DownloadFileInSegments(url, destPath, segments)

	assert.NoError(t, err)
}

func TestFailedSegmentDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := clients.NewMockHttpClient(ctrl)
	mockFactory := NewMockSegmentManagerFactory(ctrl)
	mockSegmentManager := NewMockSegmentManager(ctrl)

	url := "http://example.com/file.zip"
	destPath := "/tmp/file.zip"
	segments := 3

	// Mock the HEAD request to get file size
	mockResp := &http.Response{
		StatusCode:    http.StatusOK,
		ContentLength: 900, // Let's assume each segment is 300 bytes
		Body:          http.NoBody,
	}
	mockClient.EXPECT().Head(url).Return(mockResp, nil)

	// Mock the factory to return our mock segment manager
	mockFactory.EXPECT().NewSegmentManager(mockClient, url, destPath).Return(mockSegmentManager)

	// Mock the segment downloads
	// Let's assume the first segment download fails
	mockSegmentManager.EXPECT().DownloadSegment(gomock.Any(), mockClient, url, int64(0), int64(299), gomock.Any()).Return(fmt.Errorf("failed to download segment"))
	mockSegmentManager.EXPECT().DownloadSegment(gomock.Any(), mockClient, url, int64(300), int64(599), gomock.Any()).Return(nil)
	mockSegmentManager.EXPECT().DownloadSegment(gomock.Any(), mockClient, url, int64(600), int64(899), gomock.Any()).Return(nil)

	// Since the first segment download fails, we don't expect the other segments to be downloaded or merged
	// Thus, we don't mock expectations for them

	downloader := NewSegmentedDownloader(mockClient, mockFactory, url, destPath)
	err := downloader.DownloadFileInSegments(url, destPath, segments)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "one or more segment downloads failed. Please retry")
}

func TestHEADRequestFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := clients.NewMockHttpClient(ctrl)
	downloader := &SegmentedDownloader{Client: mockHttpClient}

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

func TestMergeSegments_MismatchedSegmentData(t *testing.T) {
	destPath := "/tmp/mergedFileMismatched"
	segmentCount := 3

	// Create segment files, but the second segment has unexpected data
	ioutil.WriteFile(fmt.Sprintf("%s.part1", destPath), []byte("segment1 data"), 0644)
	ioutil.WriteFile(fmt.Sprintf("%s.part2", destPath), []byte("unexpected data"), 0644)
	ioutil.WriteFile(fmt.Sprintf("%s.part3", destPath), []byte("segment3 data"), 0644)

	err := mergeSegments(destPath, segmentCount)
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}

	mergedData, _ := ioutil.ReadFile(destPath)
	expectedData := "segment1 dataunexpected datasegment3 data"
	if string(mergedData) != expectedData {
		t.Errorf("Segment merging failed. Expected %s but got %s", expectedData, string(mergedData))
	}

	// Cleanup
	for i := 1; i <= segmentCount; i++ {
		os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}
	os.Remove(destPath)
}

func TestMergeSegments_LargeNumberOfSegments(t *testing.T) {
	destPath := "/tmp/mergedFileLarge"
	segmentCount := 1000

	// Create a large number of segment files
	for i := 1; i <= segmentCount; i++ {
		ioutil.WriteFile(fmt.Sprintf("%s.part%d", destPath, i), []byte(fmt.Sprintf("segment%d data", i)), 0644)
		defer os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}

	err := mergeSegments(destPath, segmentCount)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Cleanup the merged file
	defer os.Remove(destPath)
}

func TestMergeSegments_ZeroByteSegments(t *testing.T) {
	destPath := "/tmp/mergedFileZeroByte"
	segmentCount := 3

	// Create segment files, but the second segment is empty
	ioutil.WriteFile(fmt.Sprintf("%s.part1", destPath), []byte("segment1 data"), 0644)
	ioutil.WriteFile(fmt.Sprintf("%s.part2", destPath), []byte(""), 0644)
	ioutil.WriteFile(fmt.Sprintf("%s.part3", destPath), []byte("segment3 data"), 0644)

	err := mergeSegments(destPath, segmentCount)
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}

	mergedData, _ := ioutil.ReadFile(destPath)
	expectedData := "segment1 datasegment3 data"
	if string(mergedData) != expectedData {
		t.Errorf("Segment merging failed. Expected %s but got %s", expectedData, string(mergedData))
	}

	// Cleanup
	for i := 1; i <= segmentCount; i++ {
		os.Remove(fmt.Sprintf("%s.part%d", destPath, i))
	}
	os.Remove(destPath)
}
