package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func setLogger(path string) *os.File {
    f, err := os.OpenFile(path, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0664)
    if err != nil {
        log.Fatal("failed to set up logger")
    }

    log.SetOutput(f)
    return f
}

func main() {
    path := os.Getenv("LOGS_PATH")
    if path == "" {
        path = "./logs"
    }
    defer setLogger(path).Close()

    router := gin.Default()
    router.POST("/upload", upload)

    router.DELETE("/rollback", rollback)

    router.GET("/files", getFiles)

    router.GET("/files/:name", getFileByName)

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "healthy")
    })

    router.Run(fmt.Sprintf(":808%s", "0"))
}
