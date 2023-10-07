package downloader

import (
	"net/http"
)

// HttpClient is an interface that wraps the Get method.
type HttpClient interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}

// RealHttpClient is a real implementation that uses http.Client.
type RealHttpClient struct{}

func (c *RealHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func (c *RealHttpClient) Head(url string) (*http.Response, error) {
	return http.Head(url)
}
