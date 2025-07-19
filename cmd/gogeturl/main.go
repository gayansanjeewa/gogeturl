package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.LoadHTMLGlob("./cmd/templates/*")
	slog.Info("Starting the server", "port", 8080)

	router.GET("/", func(context *gin.Context) {
		slog.Info("Rendering index template")
		context.HTML(http.StatusOK, "index.html", nil)
	})

	if err := router.Run(":8080"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
