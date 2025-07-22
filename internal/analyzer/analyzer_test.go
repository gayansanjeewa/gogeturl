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
	mockHTML := "<html><head><title>Welcome!</title></head><body>Hello</body></html>"

	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockHTML)),
			}, nil
		},
	}

	newAnalyzer := NewAnalyzer(mockClient)
	body, err := newAnalyzer.FetchHTML("http://example.com")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(body, "Welcome") {
		t.Errorf("Expected body to contain title, got: %s", body)
	}
}

func TestExtractTitle(t *testing.T) {
	mockHTML := "<html><head><title>Welcome!</title></head><body>Hello</body></html>"

	newAnalyzer := NewAnalyzer(nil)
	title := newAnalyzer.ExtractTitle(mockHTML)

	if title != "Welcome!" {
		t.Errorf("Expected title 'Welcome!', got '%s'", title)
	}
}
