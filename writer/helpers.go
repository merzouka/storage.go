package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/merzouka/storage.go/writer/models"
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

func getOriginal(name string) string {
    parts := strings.Split(name, ".")
    ext := ""
    if len(parts) > 1 {
        ext = fmt.Sprintf(".%s", parts[1])
    }

    parts = strings.Split(parts[0], "#")
    name = parts[0]
    name += ext
    return name
}

func getFileMetadata(file string) map[string]string {
    return map[string]string{
        "name": file,
    }
}

func getFileInfo(info os.FileInfo) map[string]string {
    return map[string]string{
        "name": info.Name(),
        "creation": info.ModTime().String(),
    }
}

func validFileName(file string) bool {
    _, err := os.Stat(fmt.Sprintf("./files/%s", file))
    return !strings.Contains(file, "/") && err == nil
}

func getFileContents(contents []byte) string {
    return string(contents)
}

func getMetadata() (map[string]string, error) {
    return nil, nil
}

func parseMetadata(metadata string) map[string]string {
    infos := strings.Split(metadata, ",")
    result := map[string]string{}
    for _, info := range infos {
        parts := strings.Split(info, "=")
        result[parts[0]] = parts[1]
    }

    return result
}

func updateOrInsertFileMetadata(file string, metadata string) map[string]interface{} {
    db := models.GetConn()
    var fileMetadata models.Metadata
    db.Where("name = ?", file).First(&fileMetadata)
    fileMetadata = models.Metadata{
        ID: fileMetadata.ID,
        Name: file,
        Metadata: metadata,
    }
    db.Save(&fileMetadata)

    return map[string]interface{}{
        "name": file,
        "metadata": parseMetadata(metadata),
    }
}
