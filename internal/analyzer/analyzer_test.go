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

func TestCountHeadings(t *testing.T) {
	mockHTML := `
		<html>
			<body>
				<h1>Main</h1>
				<h2>Sub1</h2>
				<h2>Sub2</h2>
				<h3>Detail</h3>
			</body>
		</html>`

	newAnalyzer := NewAnalyzer(nil)
	headings := newAnalyzer.CountHeadings(mockHTML)

	expected := map[string]int{
		"h1": 1,
		"h2": 2,
		"h3": 1,
	}

	for tag, count := range expected {
		if headings[tag] != count {
			t.Errorf("Expected %d for %s, got %d", count, tag, headings[tag])
		}
	}
}
