package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

func getFiles(ctx *gin.Context) {
    queryParts := []string{}
    for key, values := range ctx.Request.URL.Query() {
        for _, value := range values {
            queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
        }
    }
    query := strings.Join(queryParts, "&")

    var resp *http.Response
    var err error
    client := &http.Client{}
    for _, instance := range getInstances() {
        var req *http.Request
        req, err = http.NewRequest("GET", fmt.Sprintf("%s/files?%s", instance.Path, query), nil)
        if err != nil {
            continue
        }
        resp, err = client.Do(req)
        if err != nil || resp.StatusCode / 100 != 2 {
            continue
        }
        break
    }
    if err != nil || resp.StatusCode / 100 != 2 {
        ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch saved files",
        })
        return 
    }

    var result map[string]interface{}
    buffer := new(bytes.Buffer)
    _, err = io.Copy(buffer, resp.Body)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch saved files",
        })
        return 
    }

    json.Unmarshal(buffer.Bytes(), &result)
    ctx.JSON(http.StatusOK, result)
}

func getFileByName(ctx *gin.Context) {
    name := ctx.Param("name")
    var resp *http.Response
    var err error
    client := &http.Client{}
    for _, instance := range getInstances() {
        var req *http.Request
        req, err = http.NewRequest("GET", fmt.Sprintf("%s/files/%s", instance.Path, name), nil)
        if err != nil {
            continue
        }
        resp, err = client.Do(req)
        if err != nil || resp.StatusCode / 100 != 2 {
            continue
        }
        break
    }
    if err != nil || resp.StatusCode / 100 != 2 {
        ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch file contents",
        })
        return
    }

    buffer := new(bytes.Buffer)
    _, err = io.Copy(buffer, resp.Body)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch file contents",
        })
        return
    }

    var result map[string]interface{}
    json.Unmarshal(buffer.Bytes(), &result)
    ctx.JSON(http.StatusOK, result)
}

func uploadFiles(ctx *gin.Context) {
        form, err := ctx.MultipartForm()
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "message": "please provide valid data",
            })
            return
        }

        metadata := ctx.PostForm("meta-data")

        failedSaves := map[string]FileSaveStatus{}
        requestGroups := map[string][]map[Instance]*http.Request{}
        instances := getInstances()

        for key, files := range form.File {
            for _, file := range files {
                body := new(bytes.Buffer)
                writer := multipart.NewWriter(body)
                writer.WriteField("name", getName(key))
                fileWriter, err := writer.CreateFormFile("file", key)

                if err != nil {
                    failedSaves[key] = FileSaveStatus{
                        Status: REQUEST_FAILED,
                    }
                    break
                }

                content, err := file.Open()
                if err != nil {
                    failedSaves[key] = FileSaveStatus{
                        Status: REQUEST_FAILED,
                    }
                    break
                }

                buffer := bytes.NewBuffer(nil)
                io.Copy(buffer, content)
                fileWriter.Write(buffer.Bytes())

                err = writer.Close()
                if err != nil {
                    failedSaves[key] = FileSaveStatus{
                        Status: REQUEST_FAILED,
                    }
                    break
                }

                requests := []map[Instance]*http.Request{}
                skipGroup := false

                for _, instance := range instances {
                    request, err := http.NewRequest("POST", fmt.Sprintf("%s/upload", instance.Path), bytes.NewReader(body.Bytes()))
                    if err != nil {
                    failedSaves[key] = FileSaveStatus{
                            Status: REQUEST_FAILED,
                    }
                        skipGroup = true
                        break
                    }
                    request.Header.Add("Content-Type", writer.FormDataContentType())
                    requests = append(requests, map[Instance]*http.Request{
                        instance: request,
                    })
                }

                if !skipGroup {
                    requestGroups[key] = requests
                }
            }
        }

        for key, requestGroup := range requestGroups {
            result, err := sendGroup(key, requestGroup, metadata)
            if err != nil {
                if result.Status != SUCCESS {
                    failedSaves[key] = result
                }
            }
        }

        
        if len(failedSaves) > 0 {
            ctx.JSON(http.StatusInternalServerError, failedSaves)
        } else {
            ctx.JSON(http.StatusOK, map[string]string{
                "message": "files saved successfully",
            })
        }

    }
