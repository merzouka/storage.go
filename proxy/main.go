package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const MAX_ATTEMPTS int = 4

func send() error {

    return nil
}

func main() {
    router := gin.Default()
    router.Use(cors.Default())

    router.POST("/upload", func(ctx *gin.Context) {
        form, err := ctx.MultipartForm()
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "message": "please provide valid data",
            })
        }

        filesArr := form.File
        var failed []string
        toProcess := []string{}
        for key := range filesArr {
            toProcess = append(toProcess, key)
        }
        attempts := 0

        for {
            failed = []string{}
            for _, key := range toProcess {
                files := filesArr[key]
                for _, file := range files {
                    body := new(bytes.Buffer)
                    writer := multipart.NewWriter(body)
                    writer.WriteField("name", key)
                    fileWriter, err := writer.CreateFormFile("file", key)

                    if err != nil {
                        failed = append(failed, key)
                        break
                    }

                    content, err := file.Open()
                    if err != nil {
                        failed = append(failed, key)
                        break
                    }

                    buffer := bytes.NewBuffer(nil)
                    io.Copy(buffer, content)
                    res := buffer.Bytes()
                    fileWriter.Write(res)

                    err = writer.Close()
                    if err != nil {
                        failed = append(failed, key)
                        break
                    }

                    request, err := http.NewRequest("POST", "http://localhost:8081", body)
                    if err != nil {
                        failed = append(failed, key)
                        break
                    }
                    request.Header.Add("Content-Type", writer.FormDataContentType())

                    client := &http.Client{}
                    resp, err := client.Do(request)
                    if err != nil || (resp.StatusCode / 100 != 2) {
                        failed = append(failed, key)
                        break
                    }

                }
            }

            attempts++
            if len(failed) == 0 || attempts == MAX_ATTEMPTS {
                break
            }
            toProcess = failed
        }
        if attempts == MAX_ATTEMPTS && len(failed) > 0 {
            ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
                "error": map[string]interface{}{
                    "message": "failed to upload files",
                    "files": failed,
                },
            })
        } else {
            ctx.JSON(http.StatusOK, map[string]string{
                "message": "file saved successfully",
            })
        }

    })

    router.Run(":8080")
}
