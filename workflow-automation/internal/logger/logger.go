package logger

import (
    "fmt"
    "time"
)

func Info(slug, msg string) {
    log("INFO", slug, msg)
}

func Debug(slug, msg string) {
    log("DEBUG", slug, msg)
}

func Warn(slug, msg string) {
    log("WARN", slug, msg)
}

func Error(slug, msg string) {
    log("ERROR", slug, msg)
}

func log(level, slug, msg string) {
    timestamp := time.Now().UTC().Format(time.RFC3339Nano)
    fmt.Printf("[%s] [%s] [%s] %s\n", timestamp, level, slug, msg)
}
