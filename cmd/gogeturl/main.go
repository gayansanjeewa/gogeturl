package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gayansanjeewa/gogeturl/internal/analyzer"
	"github.com/gayansanjeewa/gogeturl/internal/handler"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort = 8080 // TODO: get from environment variables
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Static("/static", "./static")

	router.Use(func(context *gin.Context) {
		slog.Info("Incoming request", "method", context.Request.Method, "path", context.Request.URL.Path)
		context.Next()
	})

	path, _ := filepath.Abs("./cmd/templates/*")
	router.LoadHTMLGlob(path)

	slog.Info(fmt.Sprintf("Starting the server at: http://localhost:%d", defaultPort))

	analyser := analyzer.NewAnalyzer(nil)
	router.POST("/analyze", handler.AnalyzeHandler(analyser)) // TODO: Fix bug - upon POSTing form navigate to /analyze route

	router.GET("/", func(context *gin.Context) {
		slog.Info("Rendering index template")
		context.HTML(http.StatusOK, "index.html", nil)
	})

	if err := router.Run(fmt.Sprintf(":%d", defaultPort)); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
