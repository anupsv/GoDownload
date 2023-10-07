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

	_, err := provider.GetURLs()
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
}
