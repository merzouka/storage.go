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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/merzouka/storage.go/proxy/models"
)

const (
    DATABASE_CONNECTION_ERROR = "connection to database failed"
    METADATA_SAVE_ERROR = "failed to save meta-data"
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

func getInstances() []Instance {
    // url := fmt.Sprintf("%s.%s.svc.cluster.local:8080/healthy", os.Getenv("SERVICE"), os.Getenv("NAMESPACE"))
    url := "localhost:8081/healthy"
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Println("request creation failed")
        return nil
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil || resp.StatusCode / 100 != 2 {
        log.Println("failed to fetch instances")
        return nil
    }
    defer resp.Body.Close()
    var result struct {
        Result []Instance
    }

    buffer := new(bytes.Buffer)
    io.Copy(buffer, resp.Body)
    json.Unmarshal(buffer.Bytes(), &result)

    return result.Result
}

func getFileName(resp string) string {
    var result map[string]string
    json.Unmarshal([]byte(resp), &result)
    return result["name"]
}


func parseMetadata(metadata string) map[string]string {
    if metadata == "" {
        return map[string]string{}
    }
    infos := strings.Split(metadata, ",")
    result := map[string]string{}
    for _, info := range infos {
        parts := strings.Split(info, "=")
        result[parts[0]] = parts[1]
    }

    return result
}

func saveMetadata(name string, metadata string) (map[string]interface{}, error) {
    db := models.GetConn()
    if db == nil {
        log.Println(DATABASE_CONNECTION_ERROR)
        return nil, errors.New(METADATA_SAVE_ERROR)
    }
    var file models.File
    db.Where("name = ?", name).First(&file)
    file.Metadata = metadata
    file.Name = name
    result := db.Save(&file)
    if result.Error != nil {
        log.Println(fmt.Printf("failed to save meta-data for file %s", name))
    }
    return map[string]interface{}{
        "name": name,
        "metadata": parseMetadata(metadata),
    }, nil
}

