package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()
    router.POST("/", func(ctx *gin.Context) {
        buf := new(strings.Builder)
        file, _, err := ctx.Request.FormFile("file")
        name := ctx.PostForm("name")
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "message": "unable to open file, make sure to include a file in the request",
            })
            return
        }
        defer file.Close()
        _, err = io.Copy(buf, file)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "message": "unable to get the file's contents",
            })
        }
        os.WriteFile(fmt.Sprintf("./files/%s", name), []byte(buf.String()), 0644)
        ctx.JSON(http.StatusOK, map[string]string{
            "message": "file saved successfully",
        })
    })
    router.Run(":8080")
}
