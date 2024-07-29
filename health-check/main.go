package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/merzouka/storage.go/health-check/models"
	"gorm.io/gorm"
)

// constants
const PING_TIMEOUT time.Duration = 8 * time.Second
const MAX_ATTEMPTS int = 4

var failures map[uint]int = map[uint]int{}
var db *gorm.DB

func ping(instances []models.Instance, index int) {
    instance := instances[index]
    var err error
    if !instance.Healthy {
        if failures[instance.ID] > 0 {
            failures[instance.ID]--
        } else if failures[instance.ID] == 0 {
            failures[instance.ID] = MAX_ATTEMPTS
            _, err = http.Get(fmt.Sprintf("%s/ping", instance.Path))
        }
    }

    if instance.Healthy {
        _, err = http.Get(fmt.Sprintf("%s/ping", instance.Path))
    }

    if err != nil {
        failures[instance.ID]++
        if failures[instance.ID] == MAX_ATTEMPTS && instance.Healthy {
            instance.Healthy = false
            instances[index].Healthy = false
            db.Save(instance)
        }
    }
}

func checker(instances []models.Instance) {
    result := db.Where("1 = 1").Delete(&models.Instance{})
    if result.Error != nil {
        log.Fatal("failed to delete old instances")
    }
    result = db.Save(&instances)
    if result.Error != nil {
        log.Fatal("failed to save instances to database")
    }

    for _, instance := range instances {
        failures[instance.ID] = 0
    }

    for {
        for i := range instances {
            go ping(instances, i)
        }
        time.Sleep(PING_TIMEOUT)
    }
}

func getInstances(names string, namespace string) []models.Instance {
    result := []models.Instance{}
    for i, name := range strings.Split(names, ",") {
        result = append(result, models.Instance{
            ID: uint(i + 1),
            Name: name,
            Healthy: true,
            Path: fmt.Sprintf("http://%s.%s.svc.cluster.local:8080", name, namespace),
        })
    }
    return result
}

func setLogger(path string) *os.File {
    f, err := os.OpenFile(path, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0664)
    if err != nil {
        log.Fatal("failed to setup logger")
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

    db = models.GetConn()
    defer models.CloseConn()

    go checker(getInstances(os.Getenv("INSTANCES"), os.Getenv("NAMESPACE")))

    router := gin.Default()
    router.GET("/healthy", func(ctx *gin.Context) {
        var instances []models.Instance
        db.Where("healthy = ?", true).Find(&instances)
        ctx.JSON(http.StatusOK, map[string][]models.Instance{
            "result": instances,
        })
    })
    router.Run(":8080")
}
