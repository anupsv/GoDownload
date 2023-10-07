package downloader

import (
	"GoDownload/helpers"
	"context"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
)

// DownloaderInterface is an interface for the Downloader.
type DownloaderInterface interface {
	DownloadFile(url string, destPath string, bar *pb.ProgressBar, ctx context.Context) error
	DownloadFiles(provider URLProvider, dir string, threads int, ctx context.Context) []*pb.ProgressBar
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

func (d *Downloader) DownloadFile(url string, destPath string, bar *pb.ProgressBar, ctx context.Context) error {

	sugar, ok := ctx.Value("sugar").(*zap.SugaredLogger)
	if !ok {
		panic("error getting logger")
	}

	// Check if file already exists
	if _, err := os.Stat(destPath); err == nil {
		sugar.Errorw("File %s already exists. Skipping download.", "destPath", destPath)
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

func (d *Downloader) DownloadFiles(provider URLProvider, dir string, threads int, logCtx context.Context) []*pb.ProgressBar {
	urls, getUrlErr := provider.GetURLs()
	sugar, ok := logCtx.Value("sugar").(*zap.SugaredLogger)
	if !ok {
		panic("error getting logger")
	}

	if getUrlErr != nil {
		sugar.Errorw("Failed to execute someFunction", "error", getUrlErr)
	}

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
		sugar.Errorw("Error starting progress pool:", "err", err)
		return nil
	}
	defer pool.Stop()

	bars := make([]*pb.ProgressBar, len(urls))
	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)

	for i, eachUrl := range urls {
		resp, respErr := d.Client.Head(eachUrl)
		if respErr != nil {
			sugar.Errorw("Error making HEAD request", "url", eachUrl, "err", err)
			continue
		}
		contentLength := resp.ContentLength
		resp.Body.Close()

		bars[i] = pb.New64(contentLength)
		pool.Add(bars[i])

		wg.Add(1)
		go func(url string, bar *pb.ProgressBar, logCtx context.Context) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				bar.Finish()
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			destPath := path.Join(dir, helpers.GetFileNameFromURL(url))
			if _, pathErr := os.Stat(destPath); os.IsNotExist(pathErr) {
				downloadErr := d.DownloadFile(url, destPath, bar, logCtx)
				if downloadErr != nil {
					fmt.Printf("Error downloading %s: %v\n", url, downloadErr)
					sugar.Errorw("Error downloading", "url", url, "err", downloadErr)
				} else {
					sugar.Infow("URL Downloaded", "url", url, "destPath", destPath)
				}
			}
			bar.Finish()
		}(eachUrl, bars[i], logCtx)
	}

	wg.Wait()
	return bars
}
