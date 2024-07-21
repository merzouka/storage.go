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

func main() {
    number := os.Getenv("SERVER_NUMBER")
    if number == "" {
        number = "1"
    }

    router := gin.Default()
    router.POST("/upload", func(ctx *gin.Context) {
        buf := new(strings.Builder)
        file, _, err := ctx.Request.FormFile("file")
        name := getName(ctx.PostForm("name"))
        metadata := ctx.PostForm("meta-data")

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
        if !validFileName(name) {
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

        log.Printf("rolling back: %s\n", name)
        ctx.String(http.StatusOK, "success")
    })

    router.GET("/files", func(ctx *gin.Context) {
        files, err := os.ReadDir("./files")
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": "failed to retrieve files",
            })
            return
        }
        fileInfos := map[string][]map[string]string{}
        for _, file := range files {
            info, err := file.Info()
            if err != nil {
                log.Println(fmt.Sprintf("failed to retrieve info for file: %s", file.Name()))
                continue
            }
            fileInfos[getOriginal(file.Name())] = append(fileInfos[getOriginal(file.Name())], getFileInfo(info))
        }
        result := []map[string]interface{}{}
        for file, revisions := range fileInfos {
            result = append(result, map[string]interface{}{
                "name": file,
                "metadata": getFileMetadata(file),
                "revisions": revisions,
            })
        }
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "result": result,
        })
    })

    router.GET("/files/:name", func(ctx *gin.Context) {
        name := ctx.Params.ByName("name")
        if !validFileName(name) {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "please provide a valid file name",
            })
            return
        }
        contents, err := os.ReadFile(fmt.Sprintf("./files/%s", name))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "please provide a valid file name",
            })
            return
        }
        ctx.JSON(http.StatusOK, map[string]string{
            "contents": getFileContents(contents),
        })
    })

    router.GET("/meta-data", func(ctx *gin.Context) {
        metadata, err := getMetadata()
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string {
                "error": "failed to retrieve meta-data",
            })
            return
        }
        ctx.JSON(http.StatusOK, metadata)
    })  

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "healthy")
    })

    router.Run(fmt.Sprintf(":808%s", number))
}
