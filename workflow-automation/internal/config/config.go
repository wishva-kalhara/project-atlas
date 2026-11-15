package config

import (
    "os"
    "path/filepath"

    "github.com/joho/godotenv"
)

type Config struct {
    WorkspaceRoot          string
    FbReelsDataPath        string
    CookiesPath            string
    GoogleFlowURL          string
    YouTubeURL             string
    FbBaseURL              string
    FbScheduledContentURL  string
    GoogleFlowEmail        string
    GoogleFlowPassword     string
}

func Load() (*Config, error) {
    // Load .env if present (ignore errors)
    _ = godotenv.Load(".env")

    // Determine workspace root (current working dir or its parent if .data-workflow-fb-schedule-in missing)
    execDir, err := os.Getwd()
    if err != nil {
        return nil, err
    }
    workspaceRoot := execDir
    if _, err := os.Stat(filepath.Join(execDir, ".data-workflow-fb-schedule-in")); os.IsNotExist(err) {
        // Running from a subdirectory, step up one level
        workspaceRoot = filepath.Clean(filepath.Join(execDir, ".."))
    }

    fbDataPath := filepath.Join(workspaceRoot, ".data-workflow-fb-schedule-in", "metadata.json")
    cookiePath := filepath.Join(workspaceRoot, "cookies.json")

    cfg := &Config{
        WorkspaceRoot:          workspaceRoot,
        FbReelsDataPath:        fbDataPath,
        CookiesPath:            cookiePath,
        GoogleFlowURL:          "https://flow.google.com/",
        YouTubeURL:             "https://www.youtube.com/",
        FbBaseURL:              "https://www.facebook.com/",
        FbScheduledContentURL:  "https://www.facebook.com/professional_dashboard/content/content_library/?filter=SCHEDULED",
        GoogleFlowEmail:        os.Getenv("GOOGLE_FLOW_EMAIL"),
        GoogleFlowPassword:     os.Getenv("GOOGLE_FLOW_PASSWORD"),
    }
    return cfg, nil
}
