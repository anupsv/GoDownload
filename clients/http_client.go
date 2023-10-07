package clients

import (
	"context"
	"net/http"
)

// HttpClient is an interface that wraps the Get method.
type HttpClient interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// RealHttpClient is a real implementation that uses http.Client.
type RealHttpClient struct{}

func (c *RealHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func (c *RealHttpClient) Head(url string) (*http.Response, error) {
	return http.Head(url)
}

func (c *RealHttpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	req = req.WithContext(ctx)
	return client.Do(req)
}
