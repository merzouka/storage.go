package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/merzouka/storage.go/writer/models"
)

func main() {
    number := os.Getenv("SERVER_NUMBER")
    models.GetConn()
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
        resp := map[string]interface{}{}
        if metadata != "" {
            resp = saveFileMetadata(getOriginal(name), metadata)
        }
        resp["name"] = name
        ctx.JSON(http.StatusOK, resp)
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

        if len(getRevisions(getOriginal(name))) == 0 {
            err := removeMetadata(getOriginal(name))
            if err != nil {
                ctx.JSON(http.StatusInternalServerError, map[string]string{
                    "message": "deleted file successfully",
                    "error": "failed to delete file meta-data",
                })
                return
            }
        }

        log.Printf("rolling back: %s\n", name)
        ctx.JSON(http.StatusOK, map[string]string{
            "message": "rollback successful",
        })
    })

    router.GET("/files", func(ctx *gin.Context) {
        log.Println("entering handler")
        files, err := os.ReadDir("./files")
        query := ctx.Query("query")
        metadataQuery := ctx.Query("meta-data")

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

            if query != "" {
                match, err := regexp.MatchString(query, file.Name())
                if err == nil && match {
                    fileInfos[getOriginal(file.Name())] = append(fileInfos[getOriginal(file.Name())], getFileInfo(info))
                }
            } else {
                fileInfos[getOriginal(file.Name())] = append(fileInfos[getOriginal(file.Name())], getFileInfo(info))
            }
        }

        result := []map[string]interface{}{}
        for file, revisions := range fileInfos {
            fileMetadata, err := getFileMetadata(file)
            if err == nil {
                if !metadataMatch(fileMetadata, metadataQuery) {
                    continue
                }
            } else {
                if metadataQuery != "" {
                    continue
                }
                result = append(result, map[string]interface{}{
                    "name": file,
                    "revisions": revisions,
                })
                continue
            }

            result = append(result, map[string]interface{}{
                "name": file,
                "metadata": parseMetadata(fileMetadata),
                "revisions": revisions,
            })
            log.Println("match")
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

        contents, err := os.ReadFile(getFilePath(name))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "please provide a valid file name",
            })
            return
        }
        metadata, err := getFileMetadata(getOriginal(name))

        ctx.JSON(http.StatusOK, map[string]interface{}{
            "contents": getFileContents(contents),
            "metadata": parseMetadata(metadata),
        })
    })

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "healthy")
    })

    router.Run(fmt.Sprintf(":808%s", number))
}
