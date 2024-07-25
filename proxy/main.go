package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const MAX_ATTEMPTS int = 4

type FileSaveStatus struct {
    Status string
    Instances []string
}

type Instance struct {
    Name string
    Path string
}

const (
    REQUEST_FAILED = "REQUEST_FAILED"
    ROLLBACK_ERROR = "ROLLBACK_ERROR"
    SAVE_ERROR = "SAVE_ERROR"
    SUCCESS = "SUCCESS"
)

func getTag() string {
    hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().Unix() + rand.Int63())))
    return hex.EncodeToString(hash[:])
}

func getName(original string) string {
    parts := strings.Split(original, ".")
    name := parts[0]
    name += "#" + getTag()
    if len(parts) > 1 {
        name += fmt.Sprintf(".%s", parts[1])
    }
    return name
}

func getInstances()[]Instance {
    return []Instance{
        {
            Name: "instance1",
            Path: "http://localhost:8081",
        },
        // {
        //     Name: "instance2",
        //     Path: "http://localhost:8082",
        // },
    }
}

func getFileName(resp string) string {
    var result map[string]string
    json.Unmarshal([]byte(resp), &result)
    return result["name"]
}

func send(instance Instance, request *http.Request, resolved chan map[Instance]string) error {
    failure := true
    var resp *http.Response = &http.Response{}
    var err error
    for i := 0; i < MAX_ATTEMPTS; i++ {
        client := &http.Client{}
        resp, err = client.Do(request)
        if err != nil {
            continue
        }
        if resp.StatusCode / 100 != 2 {
            io.ReadAll(resp.Body)
            resp.Body.Close()
            continue
        }
        failure = false
        break
    }

    if resp != nil && resp.Body != nil {
        defer resp.Body.Close()
    }

    if !failure {
        resultStr := new(strings.Builder)
        io.Copy(resultStr, resp.Body)
        name := getFileName(resultStr.String())
        resolved <- map[Instance]string{
            instance: name,
        }
        return errors.New("failed to send request to back-end")
    }

    resolved <- map[Instance]string{
        instance: "",
    }
    return nil
}

func sendGroup(key string, requestGroup []map[Instance]*http.Request, metadata string) (FileSaveStatus, error) {
    resolvedInstanceRequests := make(chan map[Instance]string)
    for _, instanceRequest := range requestGroup {
        for instance, request := range instanceRequest {
            go send(instance, request, resolvedInstanceRequests)
        }
    }

    // checking result
    succeeded := map[Instance]string{}
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
                failed = append(failed, instance.Name)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
            body := bytes.NewBuffer(content)
            req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/rollback", instance.Path), body)
            if err != nil {
                failed = append(failed, instance.Name)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
            resp, err := client.Do(req)
            if err != nil || resp.StatusCode / 100 != 2 {
                failed = append(failed, instance.Name)
                log.Println(fmt.Sprintf("failed to rollback %s", key))
                continue
            }
        }

        if len(failed) > 0 {
            return FileSaveStatus{
                Status: ROLLBACK_ERROR,
                Instances: failed,
            }, errors.New(fmt.Sprintf("rollback failed for file %s, failed instances: %s", key, strings.Join(failed, ", ")))
        } else {
            return FileSaveStatus{
                Status: SAVE_ERROR,
                Instances: nil,
            }, errors.New(fmt.Sprintf("failed to save file %s", key))
        }
    }

    _, err := saveMetadata(key, metadata)
    if err != nil {
        return FileSaveStatus{
            Status: METADATA_SAVE_ERROR,
            Instances: nil,
        }, errors.New(fmt.Sprintf("failed to save meta-data for file %s", key))
    }

    return FileSaveStatus{
        Status: SUCCESS,
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

    })

    router.GET("/files", func(ctx *gin.Context) {
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
    })

    router.GET("/files/:name", func(ctx *gin.Context) {
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
    })

    router.Run(":8080")
}
