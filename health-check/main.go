package main

import (
	"context"
	"fmt"
	"net/http"
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

func checker() {
    var instances []models.Instance
    db.Find(&instances)

    for _, instance := range instances {
        failures[instance.ID] = 0
    }

    for {
        for i := range instances {
            go ping(instances, i)
        }
        ctx, cancel := context.WithTimeout(context.Background(), PING_TIMEOUT)
        defer cancel()
        <-ctx.Done()
    }
}

func main() {
    db = models.GetConn()
    defer models.CloseConn()

    go checker()

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
