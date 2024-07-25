package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/merzouka/storage.go/proxy/models"
)

const (
    DATABASE_CONNECTION_ERROR = "connection to database failed"
    METADATA_SAVE_ERROR = "failed to save meta-data"
)

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

