package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gayansanjeewa/gogeturl/internal/analyzer"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockAnalyzer implements the Analyzer interface with stubbed values
type mockAnalyzer struct{}

func (m *mockAnalyzer) FetchHTML(url string) (string, error) {
	return `
		<!DOCTYPE html>
		<html>
			<head><title>Mock Title</title></head>
			<body>
				<h1>Heading</h1>
				<form><input type="password" /></form>
				<a href="/internal">Internal Link</a>
				<a href="http://external.com">External Link</a>
				<a href="http://broken-link.com">Broken Link</a>
			</body>
		</html>
	`, nil
}

func (m *mockAnalyzer) DetectHTMLVersion(body string) string {
	return "HTML 5"
}

func (m *mockAnalyzer) ExtractTitle(body string) string {
	return "Mock Title"
}

func (m *mockAnalyzer) CountHeadings(body string) map[string]int {
	return map[string]int{"h1": 1}
}

func (m *mockAnalyzer) DetectLoginForm(body string) bool {
	return false
}

func (m *mockAnalyzer) AnalyzeLinks(body, baseURL string) (int, int, int, error) {
	return 1, 1, 0, nil
}

type failingMockAnalyzer struct {
	mockAnalyzer
}

func (m *failingMockAnalyzer) FetchHTML(url string) (string, error) {
	return "", fmt.Errorf("mock fetch error")
}

func setUp(a analyzer.Analyzer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	path, _ := filepath.Abs("../../cmd/templates/*")
	router.LoadHTMLGlob(path)

	router.POST("/analyze", AnalyzeHandler(a))
	return router
}

func TestAnalyzeHandler(t *testing.T) {
	mock := &mockAnalyzer{}
	router := setUp(mock)

	form := url.Values{}
	form.Add("url", "http://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	body := recorder.Body.String()
	assert.Contains(t, body, "Mock Title")
	assert.Contains(t, body, "HTML 5")
	assert.Contains(t, body, "Analyzing: http://example.com")
	assert.Contains(t, body, "h1")
	assert.Contains(t, body, "Internal Links")
	assert.Contains(t, body, "External Links")
	assert.Contains(t, body, "Broken Links")
	assert.Contains(t, body, "Login Form Detection")
}

func TestAnalyzeHandler_EmptyURL(t *testing.T) {
	mock := &mockAnalyzer{}
	router := setUp(mock)

	form := url.Values{}
	form.Add("url", "")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Please provide a URL.")
}

func TestAnalyzeHandler_InvalidURL(t *testing.T) {
	mock := &mockAnalyzer{}
	router := setUp(mock)

	form := url.Values{}
	form.Add("url", "invalid-url")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid URL format")
}

func TestAnalyzeHandler_FetchFailure(t *testing.T) {
	mock := &failingMockAnalyzer{}
	router := setUp(mock)

	form := url.Values{}
	form.Add("url", "http://example.com")
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Unable to fetch the provided URL")
}
