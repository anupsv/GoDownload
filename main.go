package main

import (
	"GoDownload/downloader"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

func main() {
	// Define flags
	helpFlag := flag.Bool("help", false, "Display help information")
	threads := flag.Int("threads", runtime.NumCPU(), "Number of threads for downloading")
	dir := flag.String("dir", "./", "Download directory")
	segments := flag.Int("segments", 1, "Number of segments for downloading (max 6). Cannot be used with -threads.")

	// Define a custom flag for multiple URLs
	var urls multiFlag
	flag.Var(&urls, "url", "URL(s) to download. Can be specified multiple times.")

	// Parse flags
	flag.Parse()

	// Validate segments and threads
	if *segments > 0 {
		if *threads != runtime.NumCPU() {
			fmt.Println("Error: Cannot specify both segments and threads.")
			return
		}
		if *segments > 6 {
			fmt.Println("Error: Maximum of 6 segments allowed.")
			return
		}
	} else {
		fmt.Println("segments was given a weird value.")
		return
	}

	factory := &downloader.RealDownloaderFactory{}
	err := RunDownloader(*helpFlag, *threads, *dir, urls, factory)
	if err != nil {
		fmt.Println(err)
	}
}

func RunDownloader(helpFlag bool, threads int, dir string, urls []string, factory downloader.DownloaderFactory) error {
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
		return fmt.Errorf("Directory %s does not exist", dir)
	}

	// Check for write permissions by attempting to create a temporary file
	tempFile, err := ioutil.TempFile(dir, "write_check_")
	if err != nil {
		return fmt.Errorf("No write permissions to directory %s", dir)
	}
	tempFile.Close()
	os.Remove(tempFile.Name()) // Cleanup the temporary file

	// Create downloader instance using the factory
	dl := factory.NewDownloader(&downloader.RealHttpClient{})

	if len(urls) == 0 {
		return fmt.Errorf("Please provide URLs to download using the -url flag.")
	}

	provider := &downloader.StaticURLProvider{URLs: urls}
	dl.DownloadFiles(provider, dir, threads)
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
