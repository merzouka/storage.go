package main

import (
	"fmt"
	"regexp"
)

func main() {
    match, err := regexp.MatchString("hello=false", "dudas=false,man=true")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(match)
}
