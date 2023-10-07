package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
)

// DownloaderInterface is an interface for the Downloader.
type DownloaderInterface interface {
	DownloadFile(url string, destPath string, bar *pb.ProgressBar) error
	DownloadFiles(provider URLProvider, dir string, threads int) []*pb.ProgressBar
}

// Ensure Downloader implements DownloaderInterface
var _ DownloaderInterface = &Downloader{}

// Downloader is responsible for downloading files.
type Downloader struct {
	Client HttpClient
}

func New(client HttpClient) *Downloader {
	return &Downloader{Client: client}
}

// DownloaderFactory is an interface for creating new Downloader instances.
type DownloaderFactory interface {
	NewDownloader(client HttpClient) DownloaderInterface
}

// RealDownloaderFactory is the real implementation of DownloaderFactory.
type RealDownloaderFactory struct{}

func (rdf *RealDownloaderFactory) NewDownloader(client HttpClient) DownloaderInterface {
	return New(client)
}

func (d *Downloader) DownloadFile(url string, destPath string, bar *pb.ProgressBar) error {

	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		fmt.Printf("File %s already exists. Skipping download.\n", destPath)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	resp, err := d.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("failed to download file: %s", resp.Status))
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	progressReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, progressReader)
	return err
}

func (d *Downloader) DownloadFiles(provider URLProvider, dir string, threads int) []*pb.ProgressBar {
	urls := provider.GetURLs()

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		cancel()
	}()

	// Create a progress pool
	pool, err := pb.StartPool()
	if err != nil {
		fmt.Println("Error starting progress pool:", err)
		return nil
	}
	defer pool.Stop()

	bars := make([]*pb.ProgressBar, len(urls))
	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)

	for i, eachUrl := range urls {
		resp, respErr := d.Client.Head(eachUrl)
		if respErr != nil {
			fmt.Printf("Error making HEAD request to %s: %v\n", eachUrl, err)
			continue
		}
		contentLength := resp.ContentLength
		resp.Body.Close()

		bars[i] = pb.New64(contentLength)
		pool.Add(bars[i])

		wg.Add(1)
		go func(url string, bar *pb.ProgressBar) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				bar.Finish()
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			destPath := path.Join(dir, getFileNameFromURL(url))
			if _, pathErr := os.Stat(destPath); os.IsNotExist(pathErr) {
				downloadErr := d.DownloadFile(url, destPath, bar)
				if downloadErr != nil {
					fmt.Printf("Error downloading %s: %v\n", url, downloadErr)
				} else {
					fmt.Printf("Downloaded %s to %s\n", url, destPath)
				}
			}
			bar.Finish()
		}(eachUrl, bars[i])
	}

	wg.Wait()
	return bars
}
