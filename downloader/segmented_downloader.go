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
	Client clients.HttpClient
}

//func NewSegmentedDownloader(client HttpClient) *SegmentedDownloader {
//	return &SegmentedDownloader{Client: client}
//}

func (d *Downloader) DownloadFileInSegments(url string, destPath string, segments int) error {
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

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			resp, err := d.Client.Do(ctx, req)
			if err != nil {
				successCh <- false
				return
			}
			defer resp.Body.Close()

			// Create a segment file
			partFile, err := os.Create(fmt.Sprintf("%s.part%d", destPath, start))
			if err != nil {
				successCh <- false
				return
			}
			defer partFile.Close()

			// Write data directly to the segment file with progress update
			buf := make([]byte, 32*1024) // 32KB buffer
			for {
				nr, er := resp.Body.Read(buf)
				if nr > 0 {
					nw, ew := partFile.Write(buf[:nr])
					if ew != nil {
						successCh <- false
						return
					}
					if nw != nr {
						successCh <- false
						return
					}
					bar.Add(nw)
				}
				if er == io.EOF {
					break
				}
				if er != nil {
					successCh <- false
					return
				}
			}
			successCh <- true
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
	err = mergeSegments(destPath, segments)
	if err != nil {
		return fmt.Errorf("error merging segments: %v", err)
	}

	return nil
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
