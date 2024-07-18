package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const MAX_ATTEMPTS int = 4

type FileSaveStatus struct {
    Status string
    Instances []string
}

func send(instance string, request *http.Request, resolved chan map[string]string) error {
    failure := false
    var resp *http.Response
    for i := 0; i < MAX_ATTEMPTS; i++ {
        client := &http.Client{}
        resp, err := client.Do(request)
        if err == nil && resp.StatusCode / 100 == 2{
            failure = true
            break
        }
    }

    if failure {
        resolved <- map[string]string{
            instance: "",
        }
        return errors.New("failed to send request to back-end")
    }

    name := new(strings.Builder)
    io.Copy(name, resp.Body)
    resolved <- map[string]string{
        instance: name.String(),
    }
    return nil
}

func sendGroup(key string, requestGroup []map[string]*http.Request, resolvedGroupRequests chan map[string]bool) (FileSaveStatus, error) {
    resolvedInstanceRequests := make(chan map[string]string)
    for _, instanceRequest := range requestGroup {
        for instance, request := range instanceRequest {
            go send(instance, request, resolvedInstanceRequests)
        }
    }

    // checking result
    succeeded := map[string]string{}
    rollback := false
    for range requestGroup {
        result := <-resolvedInstanceRequests
        for instance, name := range result {
            if len(name) == 0 {
                rollback = true
            } else {
                succeeded[instance] = name
            }
        }
    }

    // rolling back
    if rollback {
        failed := []string{}
        client := &http.Client{}
        for instance, name := range succeeded {
            content, err := json.Marshal(map[string]string{
                "name": name,
            })
            if err != nil {
                failed = append(failed, instance)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
            body := bytes.NewBuffer(content)
            req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/rollback", instance), body)
            if err != nil {
                failed = append(failed, instance)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
            resp, err := client.Do(req)
            if err != nil || resp.StatusCode / 100 != 2 {
                failed = append(failed, instance)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
        }

        resolvedGroupRequests <- map[string]bool{
            key: false,
        }
        if len(failed) > 0 {
            return FileSaveStatus{
                Status: "",
                Instances: failed,
            }, errors.New(fmt.Sprintf("rollback failed for file %s, failed instances: %s", key, strings.Join(failed, ", ")))
        } else {
            return FileSaveStatus{
                Status: "SAVE_FAILURE",
                Instances: nil,
            }, errors.New(fmt.Sprintf("failed to save file %s", key))
        }
    }

    resolvedGroupRequests <- map[string]bool{
        key: true,
    }
    return FileSaveStatus{
        Status: "SUCCESS",
        Instances: nil,
    }, nil
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
        attempts := 0
        requestGroups := map[string][]*http.Request{}
        instances := []string{
            "http://localhost:8081",
        }

        for key, files := range filesArr {
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

                requests := []*http.Request{}
                skipGroup := false
                for _, instance := range instances {
                    request, err := http.NewRequest("POST", instance, body)
                    if err != nil {
                        failed = append(failed, key)
                        skipGroup = true
                        break
                    }
                    request.Header.Add("Content-Type", writer.FormDataContentType())
                    requests = append(requests, request)
                }
                if !skipGroup {
                    requestGroups[key] = requests
                }
            }
        }

        resolved := make(chan map[string]bool)
        for key, requestGroup := range requestGroups {
        }

        for range requestGroups {
            result := <-resolved
            for key, status := range result {
                if !status {
                    failed = append(failed, key)
                }
            }
            
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
