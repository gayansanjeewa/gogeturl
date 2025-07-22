package analyzer

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestFetchHTML(t *testing.T) {
	mockHTML := "<html><head><title>Test Page</title></head><body>Hello</body></html>"

	// Replace the global client with a mock
	Client = &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockHTML)),
			}, nil
		},
	}

	body, err := FetchHTML("http://example.com")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(body, "Test Page") {
		t.Errorf("Expected body to contain title, got: %s", body)
	}
}
