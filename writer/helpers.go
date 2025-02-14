package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/merzouka/storage.go/writer/models"
)

const (
    DATABASE_CONNECTION_ERROR = "failed to connect to database"
)

func getTag() string {
    hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().Unix() + rand.Int63())))
    return hex.EncodeToString(hash[:])
}

func getName(original string) string {
    return original
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

func getFileMetadata(name string) (string, error) {
    db := models.GetConn()
    if db == nil {
        return "", errors.New(DATABASE_CONNECTION_ERROR)
    }

    var file models.File
    result := db.Where("name = ?", name).First(&file)
    if result.Error != nil {
        return "", errors.New("file has no meta-data saved")
    }

    return file.Metadata, nil
}

func getFileInfo(info os.FileInfo) map[string]string {
    return map[string]string{
        "name": info.Name(),
        "creation": info.ModTime().String(),
    }
}

func getOriginalFiles() []string {
    tmp := map[string]struct{}{}
    files, err := os.ReadDir("./files/")
    if err != nil {
        return []string{}
    }
    for _, file := range files {
        info, err := file.Info()
        if err != nil {
            continue
        }
        tmp[getOriginal(info.Name())] = struct{}{}
    }
    result := []string{}
    for name := range tmp {
        result = append(result, name)
    }
    return result
}

func validOriginalFile(file string) bool {
    log.Println(getOriginalFiles())
    return slices.Contains(getOriginalFiles(), file)
}

func validFileName(file string) bool {
    _, err := os.Stat(fmt.Sprintf("./files/%s", file))
    return !strings.Contains(file, "/") && err == nil
}

func getFileContents(contents []byte) string {
    return string(contents)
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

func saveFileMetadata(file string, metadata string) map[string]interface{} {
    db := models.GetConn()
    if db == nil {
        return map[string]interface{}{
            "message": "saved file successfully",
            "error": "failed to save file meta-data",
        }
    }
    var fileMetadata models.File
    db.Where("name = ?", file).First(&fileMetadata)
    fileMetadata = models.File{
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

func getRevisions(original string) []os.DirEntry {
    files, err := os.ReadDir("./files/")
    if err != nil {
        return []os.DirEntry{}
    }
    parts := strings.Split(original, ".")
    exp := fmt.Sprintf("^%s#[0-9a-zA-Z]+", parts[0])
    if len(parts) > 1 {
        exp += "." + parts[1]
    }
    result := []os.DirEntry{}
    for _, file := range files {
        if match, err := regexp.MatchString(exp, file.Name()); err == nil && match {
            result = append(result, file)
        }
    }

    return result
}

func removeMetadata(name string) error {
    db := models.GetConn()
    if db == nil {
        return errors.New(DATABASE_CONNECTION_ERROR)
    }

    var file models.File
    result := db.Where("name = ?", name).First(&file)
    if result.Error != nil {
        return nil
    }

    result = db.Delete(&file)
    if result.Error != nil {
        return errors.New("failed to delete file meta-data")
    }

    return nil
}

func getFilePath(name string) string {
    var latest os.FileInfo = nil
    var err error
    if !strings.Contains(name, "#") {
        revisions := getRevisions(name)
        log.Println(revisions)
        if len(revisions) == 0 {
            return ""
        }
        latest, err = revisions[0].Info()
        log.Println(err)
        if err != nil {
            return "" 
        }
        for _, file := range revisions {
            info, err := file.Info()
            if err != nil {
                continue
            }
            if info.ModTime().After(latest.ModTime()) {
                latest = info
            }
        }
        name = latest.Name()
    }
    log.Println(fmt.Sprintf("file path is %s", fmt.Sprintf("./files/%s", name)))
    return fmt.Sprintf("./files/%s", name)
}

func metadataMatch(file string, query string) bool {
    metadata := parseMetadata(query)

    for key, value := range metadata {
        exp := fmt.Sprintf("%s=%s", key, value)
        match, err := regexp.MatchString(exp, file)
        if err != nil || !match {
            return false
        }
    }

    return true
}
