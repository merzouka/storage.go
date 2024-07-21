package main

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
    "fmt"
    "strings"
    "os"
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

func getMetadata(file string) map[string]string {
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
