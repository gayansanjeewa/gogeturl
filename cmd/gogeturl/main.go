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
	slog.Info("Starting the server", "port", 8080)

	router.GET("/", func(context *gin.Context) {
		slog.Info("Listening root route")
		context.String(http.StatusOK, "Go get url! 🏃‍♂️‍➡️️")
	})

	if err := router.Run(":8080"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
