package downloader

import (
	"testing"
)

func TestStaticURLProvider_GetURLs(t *testing.T) {
	provider := &StaticURLProvider{
		URLs: []string{
			"https://www.google.com",
			"invalid-url",
		},
	}

	urls := provider.GetURLs()
	if len(urls) != 1 || urls[0] != "https://www.google.com" {
		t.Fatalf("Expected valid URLs only, got %v", urls)
	}
}

// TODO: Add tests for FileURLProvider once its implementation is provided.
