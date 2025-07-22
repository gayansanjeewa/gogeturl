package handler

import (
	"log/slog"
	"net/http"

	"github.com/gayansanjeewa/gogeturl/internal/analyzer"
	"github.com/gayansanjeewa/gogeturl/internal/utils"
	"github.com/gin-gonic/gin"
)

func AnalyzeHandler(context *gin.Context) {
	url := context.PostForm("url")

	if url == "" {
		slog.Warn("URL is missing in the form submission")
		context.HTML(http.StatusBadRequest, "index.html", gin.H{
			"Error": "Please provide a URL.",
		})
		return
	}

	if err := utils.ValidateURL(url); err != nil {
		slog.Warn("Invalid URL", "error", err)
		context.HTML(http.StatusBadRequest, "index.html", gin.H{
			"Error": err.Error(),
		})
		return
	}

	slog.Info("Received URL for analysis", "url", url)

	body, err := analyzer.FetchHTML(url)
	if err != nil {
		slog.Error("Failed to fetch HTML", "error", err)
		context.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"Error": "Failed to fetch the URL: " + err.Error(),
		})
		return
	}

	title := analyzer.ExtractTitle(body)
	headings := analyzer.CountHeadings(body)
	// NOTE: AnalyzeLinks is not updated to use string body, so we skip it for now or update it separately.

	context.HTML(http.StatusOK, "index.html", gin.H{
		"Message":  "Analyzing: " + url,
		"TitleTag": title,
		"Headings": headings,
		//"InternalLinks": internal,
		//"ExternalLinks": external,
		//"BrokenLinks":   broken,
	})
}
