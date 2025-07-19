package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnalyzeHandler handles POST /analyze
func AnalyzeHandler(context *gin.Context) {
	url := context.PostForm("url")

	if url == "" {
		slog.Warn("URL is missing in the form submission")
		context.HTML(http.StatusBadRequest, "index.html", gin.H{
			"Error": "Please provide a URL.",
		})
		return
	}

	slog.Info("Received URL for analysis", "url", url)

	// Placeholder response
	context.HTML(http.StatusOK, "index.html", gin.H{
		"Message": "Analyzing: " + url,
	})
}
