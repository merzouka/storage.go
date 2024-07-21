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
        name := new(strings.Builder)
        io.Copy(name, resp.Body)
        log.Println(name)
        resolved <- map[Instance]string{
            instance: name.String(),
        }
        return errors.New("failed to send request to back-end")
    }

    resolved <- map[Instance]string{
        instance: "",
    }
    return nil
}

func sendGroup(key string, requestGroup []map[Instance]*http.Request) (FileSaveStatus, error) {
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

        filesArr := form.File
        failedSaves := map[string]FileSaveStatus{}
        requestGroups := map[string][]map[Instance]*http.Request{}
        instances := getInstances() 

        for key, files := range filesArr {
            for _, file := range files {
                body := new(bytes.Buffer)
                writer := multipart.NewWriter(body)
                writer.WriteField("name", key)
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
            result, err := sendGroup(key, requestGroup)
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

    router.Run(":8080")
}
