package main

import (
	"GoDownload/clients"
	"GoDownload/downloader"
	"GoDownload/helpers"
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"runtime"
)

var logger *zap.Logger

func main() {

	// Initialize logger
	logger, _ = zap.NewProduction()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger) // flushes buffer, if any

	sugar := logger.Sugar()
	sugar.Infow("Logger initialized", "mode", "production")

	// Define flags
	helpFlag := flag.Bool("help", false, "Display help information")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads for downloading")
	dir := flag.String("dir", "./", "Download directory")
	segments := flag.Int("segments", 1, "Number of segments for downloading (max 6). Cannot be used with -threads.")
	ctx := context.WithValue(context.Background(), "sugar", sugar)

	// Define a custom flag for multiple URLs
	var urls multiFlag
	flag.Var(&urls, "url", "URL(s) to download. Can be specified multiple times.")

	// Parse flags
	flag.Parse()

	err := helpers.ValidateDirectory(*dir)
	if err != nil {
		sugar.Errorw("Failed ValidateDirectory call", "error", err)
		return
	}

	// Validate segments and threads
	if *segments > 0 {
		if *threads != runtime.NumCPU() {
			sugar.Errorw("Error: Cannot specify both segments and threads.")
			return
		}
		if *segments > 6 {
			sugar.Errorw("Error: Maximum of 6 segments allowed.")
			return
		}
	} else {
		sugar.Errorw("segments was given a weird value.")
		return
	}

	factory := &downloader.RealDownloaderFactory{}
	dlErr := RunDownloader(*helpFlag, *threads, *dir, urls, factory, ctx)
	if dlErr != nil {
		sugar.Errorw("Problem running downloader", dlErr)
	}
}

func RunDownloader(helpFlag bool, threads int, dir string, urls []string, factory downloader.DownloaderFactory, ctx context.Context) error {
	// Display help information if --help is provided
	if helpFlag {
		flag.PrintDefaults()
		return nil
	}

	// Limit threads to max available
	if threads > runtime.NumCPU() {
		threads = runtime.NumCPU()
	}

	// Check if number of URLs is less than the specified threads
	if len(urls) < threads {
		fmt.Printf("Warning: Number of URLs (%d) is less than the specified threads (%d). "+
			"Setting threads to %d.\n", len(urls), threads, len(urls))
		threads = len(urls)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	// Check for write permissions by attempting to create a temporary file
	tempFile, err := ioutil.TempFile(dir, "write_check_")
	if err != nil {
		return fmt.Errorf("no write permissions to directory %s", dir)
	}
	tempFile.Close()
	os.Remove(tempFile.Name()) // Cleanup the temporary file

	// Create downloader instance using the factory
	dl := factory.NewDownloader(&clients.RealHttpClient{})

	if len(urls) == 0 {
		return fmt.Errorf("please provide URLs to download using the -url flag")
	}

	provider := &clients.StaticURLProvider{URLs: urls}
	dl.DownloadFiles(provider, dir, threads, ctx)
	return nil
}

// multiFlag allows to specify a flag multiple times and collect all values into a slice.
type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprint(*m)
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}
