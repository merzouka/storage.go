package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func getName(original string) string {
    return original
}

func main() {
    number := os.Getenv("SERVER_NUMBER")
    fmt.Printf("server: %s\n", number)
    router := gin.Default()
    router.POST("/upload", func(ctx *gin.Context) {
        buf := new(strings.Builder)
        file, _, err := ctx.Request.FormFile("file")
        name := getName(ctx.PostForm("name"))
        if strings.Contains(name, number) {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": "fuck you",
            })
            log.Println("won't save")
            return
        }

        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "message": "unable to open file, make sure to include a file in the request",
            })
            log.Println("file open error")
            return
        }

        defer file.Close()
        _, err = io.Copy(buf, file)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "message": "unable to get the file's contents",
            })
            log.Println("file open error")
            return
        }

        log.Println(fmt.Sprintf("saving:  %s\n", name))
        os.WriteFile(fmt.Sprintf("./files/%s", name), []byte(buf.String()), 0644)
        ctx.String(http.StatusOK, name)
    })

    router.DELETE("/rollback", func(ctx *gin.Context) {
        bodyStr := new(strings.Builder)
        io.Copy(bodyStr, ctx.Request.Body)
        var body map[string]string
        err := json.Unmarshal([]byte(bodyStr.String()), &body)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "please provide a valid file name",
            })
            return
        }
        name := body["name"]
        if strings.Contains(name, "/") || strings.Contains(name, "3") {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "please provide a valid file name",
            })
            return
        }
        err = os.Remove(fmt.Sprintf("./files/%s", name))
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": "failed to rollback operation",
            })
            return
        }
        ctx.String(http.StatusOK, "success")
    })

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "healthy")
    })

    router.Run(fmt.Sprintf(":808%s", number))
}
