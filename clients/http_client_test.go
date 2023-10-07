package clients

import (
	"context"
	"github.com/golang/mock/gomock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealHttpClient_Get(t *testing.T) {
	client := &RealHttpClient{}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRealHttpClient_Head(t *testing.T) {
	// Create a test server that responds to HEAD requests
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("Expected HEAD request, got: %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := &RealHttpClient{}

	resp, err := client.Head(ts.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Since it's a HEAD request, the body should be empty
	if len(body) != 0 {
		t.Fatalf("Expected empty response body for HEAD request, got: %s", body)
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got: %d", resp.StatusCode)
	}
}

func TestDo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHttpClient := NewMockHttpClient(ctrl)
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// Setting up the mock expectation
	mockHttpClient.EXPECT().Do(context.Background(), req).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	// Call the Do method
	resp, err := mockHttpClient.Do(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got: %d", http.StatusOK, resp.StatusCode)
	}
}
