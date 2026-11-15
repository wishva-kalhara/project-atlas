package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	WorkspaceRoot         string
	FbReelsDataPath       string
	CookiesPath           string
	GoogleFlowURL         string
	YouTubeURL            string
	FbBaseURL             string
	FbScheduledContentURL string
	GoogleFlowEmail       string
	GoogleFlowPassword    string
}

func Load() (*Config, error) {
	// Load .env if present (ignore errors)
	_ = godotenv.Load(".env")

	// Determine executable directory (current working directory)
	execDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Workspace root: current working dir or its parent if .data-workflow-fb-schedule-in missing
	workspaceRoot := execDir
	if _, err := os.Stat(filepath.Join(execDir, ".data-workflow-fb-schedule-in")); os.IsNotExist(err) {
		// Running from a subdirectory, step up one level
		workspaceRoot = filepath.Clean(filepath.Join(execDir, ".."))
	}

	// fb scheduling data lives in the appropriate data directory (supports .dist and case variations)
	fbDataPath := ""
	// Prefer a data directory in the current working directory if it exists (e.g., when running from .dist)
	// First check for the .dist subfolder (common when running the binary directly)
	if _, err := os.Stat(filepath.Join(execDir, ".dist", ".data-workflow-fb-schedule-in", "metadata.json")); err == nil {
		fbDataPath = filepath.Join(execDir, ".dist", ".data-workflow-fb-schedule-in", "metadata.json")
	} else if _, err := os.Stat(filepath.Join(execDir, ".data-workflow-fb-schedule-in", "metadata.json")); err == nil {
		fbDataPath = filepath.Join(execDir, ".data-workflow-fb-schedule-in", "metadata.json")
	} else if _, err := os.Stat(filepath.Join(execDir, ".data-workflow-fbSchedule-in", "metadata.json")); err == nil {
		fbDataPath = filepath.Join(execDir, ".data-workflow-fbSchedule-in", "metadata.json")
	} else if _, err := os.Stat(filepath.Join(workspaceRoot, ".data-workflow-fb-schedule-in", "metadata.json")); err == nil {
		fbDataPath = filepath.Join(workspaceRoot, ".data-workflow-fb-schedule-in", "metadata.json")
	} else if _, err := os.Stat(filepath.Join(workspaceRoot, ".data-workflow-fbSchedule-in", "metadata.json")); err == nil {
		fbDataPath = filepath.Join(workspaceRoot, ".data-workflow-fbSchedule-in", "metadata.json")
	} else {
		// Default fallback (may error later if missing)
		fbDataPath = filepath.Join(workspaceRoot, ".data-workflow-fb-schedule-in", "metadata.json")
	}	// cookies.json should live next to the executable (wherever the binary runs from)
	cookiePath := filepath.Join(execDir, "cookies.json")

	cfg := &Config{
		WorkspaceRoot:         workspaceRoot,
		FbReelsDataPath:       fbDataPath,
		CookiesPath:           cookiePath,
		GoogleFlowURL:         "https://flow.google.com/",
		YouTubeURL:            "https://www.youtube.com/",
		FbBaseURL:             "https://www.facebook.com/",
		FbScheduledContentURL: "https://www.facebook.com/professional_dashboard/content/content_library/?filter=SCHEDULED",
		GoogleFlowEmail:       os.Getenv("GOOGLE_FLOW_EMAIL"),
		GoogleFlowPassword:    os.Getenv("GOOGLE_FLOW_PASSWORD"),
	}
	return cfg, nil
}
