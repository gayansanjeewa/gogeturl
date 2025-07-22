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

	doc, err := analyzer.FetchAndParse(url)
	if err != nil {
		slog.Error("Failed to fetch/parse HTML", "error", err)
		context.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"Error": "Failed to fetch or parse the URL: " + err.Error(),
		})
		return
	}

	slog.Info("HTML parsed successfully", "nodeType", doc.Type)

	title := analyzer.ExtractTitle(doc)
	headings := analyzer.CountHeadings(doc)

	internal, external, broken, err := analyzer.AnalyzeLinks(doc, url)
	if err != nil {
		slog.Warn("Link analysis failed", "error", err)
	}

	context.HTML(http.StatusOK, "index.html", gin.H{
		"Message":       "Analyzing: " + url,
		"TitleTag":      title,
		"Headings":      headings,
		"InternalLinks": internal,
		"ExternalLinks": external,
		"BrokenLinks":   broken,
	})
}
