package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProcessedCount struct {
    mutex sync.Mutex
    count int
}

var failedAttemptTracker ProcessedCount = ProcessedCount{
    count: 0,
}

var completedTracker ProcessedCount = ProcessedCount{
    count: 0,
}

var endChan chan bool = make(chan bool)

type Matrix = [][]int

func getMatrix(file string) Matrix {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(content), "\n")
	result := Matrix{}
	for _, line := range lines {
		if line == "" {
			break
		}
		ints := []int{}
		for _, value := range strings.Split(line, ",") {
			if value == "" {
				ints = append(ints, -1)
                continue
			}
			intVal, err := strconv.Atoi(value)
			if err != nil {
				log.Fatal(err)
			}
			ints = append(ints, intVal)
		}
		result = append(result, ints)
	}
	return result
}

func handleRequest(reqChan chan string, respChan chan int, matrix Matrix, handlers int) {
    end := false
    for {
        select {
        case <-endChan:
            end = true
        case request := <-reqChan:
            if request == "" {
                break
            }
            components := strings.Split(request, ",")
            row, err := strconv.Atoi(components[0])
            if err != nil {
                log.Fatal(err)
            }
            col, err := strconv.Atoi(components[1])
            if err != nil {
                log.Fatal(err)
            }
            if matrix[row][col] != -1 {
                respChan <- matrix[row][col]
                failedAttemptTracker.count = 0
            } else {
                reqChan <- request
                failedAttemptTracker.count++
                if failedAttemptTracker.count == handlers {
                    endChan <- true
                    end = true
                }
            }
        }
        if end {
            break
        }
    }
    endChan <- true
}

func processMatrix(matrix Matrix, reqChan chan string, respChan chan int, handlers int) {
    fail := false
    for i, row := range matrix {
        for j, value := range row {
            if value == -1 {
                reqChan <- fmt.Sprintf("%d,%d", i, j)
                select {
                case matrix[i][j] = <-respChan:
                    break
                case <-endChan:
                    fail = true
                }
            }
            if fail {
                break
            }
        }
        if fail {
            break
        }
    }
    if !fail {
        completedTracker.mutex.Lock()
        completedTracker.count++
        completedTracker.mutex.Unlock()
        completedTracker.mutex.Lock()
        if completedTracker.count == handlers {
            completedTracker.count--
            close(reqChan)
            close(respChan)
            endChan <- true
        }
        completedTracker.mutex.Unlock()
        fmt.Println(matrix)
    } else {
        fmt.Println("exiting matrix processor...")
        endChan <- true
    }
}

func main() {
    files := []string{"./sample1", "./sample2", "./sample3"}
    reqChan := make(chan string)
    respChan := make(chan int)
    for _, file := range files {
        matrix := getMatrix(file)
        go handleRequest(reqChan, respChan, matrix, len(files))
        go processMatrix(matrix, reqChan, respChan, len(files))
    }
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute * 2)
    defer cancel()
    defer close(endChan)
    end := false
    for {
        select {
        case <-ctx.Done():
            end = true
            break
        }
        if end {
            break
        }
    }
}
