package analyzer

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
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

	analyzer := NewAnalyzer(nil)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := analyzer.DetectLoginForm(testCase.html)
			assert.Equal(t, testCase.expected, result)
		})
	}
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

	analyzer := NewAnalyzer(mockClient)
	body, err := analyzer.FetchHTML("http://example.com")

	assert.NoError(t, err)
	assert.Contains(t, body, "Welcome!")
}

func TestExtractTitle(t *testing.T) {
	mockHTML := "<html><head><title>Welcome!</title></head><body>Hello</body></html>"

	analyzer := NewAnalyzer(nil)
	title := analyzer.ExtractTitle(mockHTML)

	assert.Equal(t, "Welcome!", title)
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

	analyzer := NewAnalyzer(nil)
	headings := analyzer.CountHeadings(mockHTML)

	expected := map[string]int{
		"h1": 1,
		"h2": 2,
		"h3": 1,
	}

	for tag, count := range expected {
		assert.Equal(t, count, headings[tag], "for heading %s", tag)
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

	analyzer := NewAnalyzer(mockClient)
	intCount, extCount, brokenCount, err := analyzer.AnalyzeLinks(mockHTML, "http://localhost")

	assert.NoError(t, err)
	assert.Equal(t, 2, intCount)
	assert.Equal(t, 2, extCount)
	assert.Equal(t, 0, brokenCount)
}

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5 doctype",
			html:     `<!DOCTYPE html><html><head></head><body></body></html>`,
			expected: "HTML 5",
		},
		{
			name:     "HTML 4.01 Transitional doctype",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"><html><head></head><body></body></html>`,
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "Unknown doctype",
			html:     `<html><head></head><body></body></html>`,
			expected: "Unknown",
		},
	}

	analyzer := NewAnalyzer(nil)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			version := analyzer.DetectHTMLVersion(testCase.html)
			assert.Equal(t, testCase.expected, version)
		})
	}
}
