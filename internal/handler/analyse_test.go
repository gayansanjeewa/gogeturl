package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockAnalyzer struct{}

func (m *mockAnalyzer) FetchHTML(url string) (string, error) {
	return "<!DOCTYPE html><html><head><title>Mock Title</title></head><body><h1>Heading</h1></body></html>", nil
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

func TestAnalyzeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	path, _ := filepath.Abs("../../cmd/templates/*")
	router.LoadHTMLGlob(path)

	mock := &mockAnalyzer{}
	router.POST("/analyze", AnalyzeHandler(mock))

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
}
