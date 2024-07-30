package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
    Name string `json:"name"`
    Path string `json:"path"`
}

const (
    REQUEST_FAILED = "REQUEST_FAILED"
    ROLLBACK_ERROR = "ROLLBACK_ERROR"
    SAVE_ERROR = "SAVE_ERROR"
    SUCCESS = "SUCCESS"
)

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

func setLogger(path string) *os.File {
    f, err := os.OpenFile(path, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0664)
    if err != nil {
        log.Fatal("failed to set log destination")
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
    router.Use(cors.Default())

    router.GET("/", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "hello storage app")
    })

    router.POST("/upload", uploadFiles)

    router.GET("/files", getFiles)

    router.GET("/files/:name", getFileByName)

    router.Run(":8080")
}
