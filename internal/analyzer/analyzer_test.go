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

func TestDetectLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "Contains login form with password input",
			html:     `<form><input type="text" name="username"><input type="password" name="pass"></form>`,
			expected: true,
		},
		{
			name:     "Contains only password field",
			html:     `<div><input type="password"></div>`,
			expected: true,
		},
		{
			name:     "Contains form but no password field",
			html:     `<form><input type="text" name="email"></form>`,
			expected: false,
		},
	}

	newAnalyzer := NewAnalyzer(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newAnalyzer.DetectLoginForm(tt.html)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
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

func TestAnalyzeLinks(t *testing.T) {
	// HTML structure:
	// <link href="/style.css" />                             --> internal
	// <link href="https://external.com/theme.css" />         --> external
	// <a href="/about">About</a>                             --> internal
	// <a href="https://external.com/page">External Page</a>  --> external
	// <a href="mailto:someone@example.com">Email</a>         --> should be skipped

	mockHTML := `
	<html>
		<head>
			<link href="/style.css" />
			<link href="https://external.com/theme.css" />
		</head>
		<body>
			<a href="/about">About</a>
			<a href="https://external.com/page">External Page</a>
			<a href="mailto:someone@example.com">Email</a>
		</body>
	</html>`

	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("OK")),
			}, nil
		},
	}

	newAnalyzer := NewAnalyzer(mockClient)
	intCount, extCount, brokenCount, err := newAnalyzer.AnalyzeLinks(mockHTML, "http://localhost")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if intCount != 2 {
		t.Errorf("Expected 2 internal links, got %d", intCount)
	}

	if extCount != 2 {
		t.Errorf("Expected 2 external links, got %d", extCount)
	}

	if brokenCount != 0 {
		t.Errorf("Expected 0 broken links, got %d", brokenCount)
	}
}
