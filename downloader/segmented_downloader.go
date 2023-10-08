package downloader

import (
	"GoDownload/clients"
	"context"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SegmentedDownloader struct {
	Client         clients.HttpClient
	SegmentManager SegmentManager
}

type SegmentManager interface {
	DownloadSegment(ctx context.Context, client clients.HttpClient, url string, start, end int64, destPath string) error
	MergeSegments(destPath string, segmentCount int) error
}

type SegmentManagerFactory interface {
	NewSegmentManager(client clients.HttpClient, url string, destPath string) SegmentManager
}

type RealSegmentManagerFactory struct{}

func (f *RealSegmentManagerFactory) NewSegmentManager(client clients.HttpClient, url string, destPath string) SegmentManager {
	return &FileSegmentManager{}
}

func NewSegmentedDownloader(client clients.HttpClient, factory SegmentManagerFactory, url string, destPath string) *SegmentedDownloader {
	manager := factory.NewSegmentManager(client, url, destPath)
	return &SegmentedDownloader{
		Client:         client,
		SegmentManager: manager,
	}
}

type FileSegmentManager struct{}

func (m *FileSegmentManager) DownloadSegment(ctx context.Context, client clients.HttpClient, url string, start, end int64, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("expected partial content status but got %s", resp.Status)
	}

	// Create a segment file
	partFile, err := os.Create(fmt.Sprintf("%s.part%d", destPath, start))
	if err != nil {
		return err
	}
	defer partFile.Close()

	// Write data directly to the segment file
	_, err = io.Copy(partFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (m *FileSegmentManager) MergeSegments(destPath string, segmentCount int) error {
	mergedFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer mergedFile.Close()

	for i := 0; i < segmentCount; i++ {
		segmentPath := fmt.Sprintf("%s.part%d", destPath, i)
		segmentFile, err := os.Open(segmentPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(mergedFile, segmentFile)
		segmentFile.Close()
		if err != nil {
			return err
		}

		// Remove the segment file after merging
		os.Remove(segmentPath)
	}

	return nil
}

func (d *SegmentedDownloader) DownloadFileInSegments(url string, destPath string, segments int) error {
	// Get the file size
	resp, err := d.Client.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file info: %s", resp.Status)
	}

	fileSize := resp.ContentLength
	segmentSize := fileSize / int64(segments)

	// Create a progress bar
	bar := pb.Start64(fileSize)
	defer bar.Finish()

	// Create a context for graceful exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle ctrl+c gracefully
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancel()
	}()

	var wg sync.WaitGroup
	successCh := make(chan bool, segments)

	for i := 0; i < segments; i++ {
		wg.Add(1)
		start := int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == segments-1 {
			end = fileSize - 1
		}

		go func(start, end int64, ctx context.Context) {
			defer wg.Done()
			nErr := d.SegmentManager.DownloadSegment(ctx, d.Client, url, start, end, destPath)
			//successCh <- (true)
			successCh <- (nErr == nil)
		}(start, end, ctx)
	}

	wg.Wait()
	close(successCh)

	allSuccessful := true
	for success := range successCh {
		if !success {
			allSuccessful = false
			break
		}
	}

	if !allSuccessful {
		return fmt.Errorf("one or more segment downloads failed. Please retry")
	}

	// Merge the downloaded segments
	return d.SegmentManager.MergeSegments(destPath, segments)
}

// mergeSegments merges the downloaded segments into a single file.
func mergeSegments(destPath string, segmentCount int) error {
	mergedFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer mergedFile.Close()

	for i := 1; i <= segmentCount; i++ {
		segmentPath := fmt.Sprintf("%s.part%d", destPath, i)
		segmentFile, err := os.Open(segmentPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(mergedFile, segmentFile)
		segmentFile.Close()
		if err != nil {
			return err
		}

		// Remove the segment file after merging
		os.Remove(segmentPath)
	}

	return nil
}
