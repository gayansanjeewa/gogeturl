package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gayansanjeewa/gogeturl/internal/analyzer"
	"github.com/gayansanjeewa/gogeturl/internal/utils"
	"github.com/gin-gonic/gin"
)

func AnalyzeHandler(analyzer analyzer.Analyzer) gin.HandlerFunc {
	return func(context *gin.Context) {
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
			context.HTML(http.StatusOK, "index.html", gin.H{
				"Error": "Unable to fetch the provided URL. Reason: " + err.Error(),
			})
			return
		}

		htmlVersion := analyzer.DetectHTMLVersion(body)
		title := analyzer.ExtractTitle(body)
		headings := analyzer.CountHeadings(body)
		hasLoginForm := analyzer.DetectLoginForm(body)

		internal, external, broken, err := analyzer.AnalyzeLinks(body, url)
		if err != nil {
			slog.Warn("Link analysis failed", "error", err)
		}

		context.HTML(http.StatusOK, "index.html", gin.H{
			"Message":       fmt.Sprintf("Analyzing: %s", url),
			"HTMLVersion":   htmlVersion,
			"TitleTag":      title,
			"Headings":      headings,
			"InternalLinks": internal,
			"ExternalLinks": external,
			"BrokenLinks":   broken,
			"HasLoginForm":  hasLoginForm,
		})
	}
}
