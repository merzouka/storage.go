package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/merzouka/storage.go/health-check/models"
	"gorm.io/gorm"
)

const PING_TIMEOUT time.Duration = 8 * time.Second

const MAX_ATTEMPTS int = 4

var failures map[uint]int = map[uint]int{}
var db *gorm.DB

func ping(instance *models.Instance) {
    _, err := http.Get(fmt.Sprintf("%s/ping", instance.Path))
    if err != nil {
        failures[instance.ID]++
        if failures[instance.ID] == MAX_ATTEMPTS {
            instance.Healthy = false
            db.Save(&instance)
        }
    }
}

func main() {
    var instances []models.Instance
    db = models.GetConn()
    defer models.CloseConn()
    db.Find(&instances)

    for _, instance := range instances {
        failures[instance.ID] = 0
    }

    for {
        for _, instance := range instances {
            go ping(&instance)
        }
        ctx, cancel := context.WithTimeout(context.Background(), PING_TIMEOUT)
        defer cancel()
        <-ctx.Done()
    }
}
