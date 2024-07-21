package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

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

func f() {
    files, err := os.ReadDir("./test/")
    if err != nil {
        log.Println("failed to read directory")
    }
    for _, file := range files {
        info, err := file.Info()
        if err != nil {
            log.Fatal(fmt.Sprintf("failed to read file %s\n", file.Name()))
        }
        fmt.Println(info.ModTime().UnixNano())
    }
}

func main() {
    f := map[string][]string{}
    f["hello"] = append(f["hello"], "hello")
    fmt.Println(f["hello"])
}
