package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PromptEntry mirrors the shape in .data-harness-gemini-in/human-prompts.json
type PromptEntry struct {
	Tag    string   `json:"tag"`
	Images []string `json:"images"`
	Text   string   `json:"text"`
}

// InputPayload mirrors the top-level JSON object in human-prompts.json
type InputPayload struct {
	Prompts []PromptEntry `json:"prompts"`
}

// MetadataEntry is one record written to .data-workflow-googleflow-in/metadata.json
type MetadataEntry struct {
	Tag         string   `json:"tag"`
	Images      []string `json:"images"`
	VideoPrompt string   `json:"videoPrompt"`
}

// SchedulingInfo mirrors the contract in workflow-automation/internal/data/models.go
type SchedulingInfo struct {
	IsSuccess bool `json:"isSuccess"`
	Tried     bool `json:"tried"`
}

// FBScheduleEntry is one record written to .data-workflow-fbSchedule-in/metadata.json
type FBScheduleEntry struct {
	Caption        string         `json:"caption"`
	VideoPath      string         `json:"videoPath"`
	SchedulingInfo SchedulingInfo `json:"schdulingInfo"`
}

// FBSchedulePayload is the top-level structure for fbSchedule metadata.json
type FBSchedulePayload struct {
	FbPageSlug string            `json:"fbPageSlug"`
	PageName   string            `json:"pageName"`
	Posts      []FBScheduleEntry `json:"posts"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no action binded")
		return
	}

	switch os.Args[1] {
	case "gen-workflow-googleflow-in":
		runGenWorkflowGoogleFlowIn()
	case "gen-workflow-fbSchedule-in":
		runGenWorkflowFBScheduleIn(os.Args[2:])
	case "clear-tmp-data":
		runClearTmpData()
	case "new-workflow-fbScheduling":
		runNewWorkflowFBScheduling()
	default:
		fmt.Println("no action binded")
	}
}

func runGenWorkflowGoogleFlowIn() {
	geminiInDir := ".data-harness-gemini-in"
	geminiOutDir := ".data-harness-gemini-out"
	googleflowOutDir := ".data-workflow-googleflow-in"

	promptsPath := filepath.Join(geminiInDir, "human-prompts.json")
	rawJSON, err := os.ReadFile(promptsPath)
	if err != nil {
		logFatal("Failed to read %s: %v", promptsPath, err)
	}

	var payload InputPayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		logFatal("Invalid JSON in %s: %v", promptsPath, err)
	}

	if err := os.MkdirAll(googleflowOutDir, 0755); err != nil {
		logFatal("Failed to create output folder %s: %v", googleflowOutDir, err)
	}

	var entries []MetadataEntry

	for _, entry := range payload.Prompts {
		if entry.Tag == "" {
			continue
		}

		mdPath := filepath.Join(geminiOutDir, entry.Tag+".md")
		mdBytes, err := os.ReadFile(mdPath)
		if err != nil {
			logFatal("Failed to read %s for tag '%s': %v", mdPath, entry.Tag, err)
		}

		mdText := string(mdBytes)

		// Extract everything from "FORMAT AND INTENT:" to the end of the file.
		const marker = "FORMAT AND INTENT:"
		idx := strings.Index(mdText, marker)
		if idx == -1 {
			logFatal("Marker '%s' not found in %s for tag '%s'", marker, mdPath, entry.Tag)
		}
		videoPrompt := strings.TrimRight(mdText[idx:], "\r\n ")

		// Resolve image references to absolute paths against the input dir.
		var absImages []string
		for _, imgRef := range entry.Images {
			candidate := filepath.Join(geminiInDir, filepath.Clean(imgRef))
			absPath, err := filepath.Abs(candidate)
			if err != nil {
				logFatal("Failed to resolve absolute path for image '%s' on tag '%s': %v", imgRef, entry.Tag, err)
			}
			if _, err := os.Stat(absPath); err != nil {
				logFatal("Image not found at '%s' for tag '%s': %v", absPath, entry.Tag, err)
			}
			absImages = append(absImages, absPath)
		}

		entries = append(entries, MetadataEntry{
			Tag:         entry.Tag,
			Images:      absImages,
			VideoPrompt: videoPrompt,
		})
		fmt.Printf("Processed tag '%s' -> %d chars videoPrompt\n", entry.Tag, len(videoPrompt))
	}

	outBytes, err := json.MarshalIndent(entries, "", "    ")
	if err != nil {
		logFatal("Failed to marshal metadata: %v", err)
	}
	outBytes = append(outBytes, '\n')

	outPath := filepath.Join(googleflowOutDir, "metadata.json")
	if err := os.WriteFile(outPath, outBytes, 0644); err != nil {
		logFatal("Failed to write %s: %v", outPath, err)
	}
	fmt.Printf("Wrote %s with %d entries\n", outPath, len(entries))
}

func runGenWorkflowFBScheduleIn(args []string) {
	fbScheduleFlags := flag.NewFlagSet("gen-workflow-fbSchedule-in", flag.ExitOnError)
	fbPageSlug := fbScheduleFlags.String("fbPageSlug", "", "Facebook page slug (required)")
	pageName := fbScheduleFlags.String("pageName", "", "Page name (required)")
	fbScheduleFlags.Parse(args)

	if *fbPageSlug == "" || *pageName == "" {
		fmt.Fprintln(os.Stderr, "Error: --fbPageSlug and --pageName are required")
		fbScheduleFlags.PrintDefaults()
		os.Exit(1)
	}

	geminiInDir := ".data-harness-gemini-in"
	geminiOutDir := ".data-harness-gemini-out"
	fbScheduleOutDir := ".data-workflow-fbSchedule-in"
	googleflowOutDir := ".data-workflow-googleflow-out"

	promptsPath := filepath.Join(geminiInDir, "human-prompts.json")
	rawJSON, err := os.ReadFile(promptsPath)
	if err != nil {
		logFatal("Failed to read %s: %v", promptsPath, err)
	}

	var payload InputPayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		logFatal("Invalid JSON in %s: %v", promptsPath, err)
	}

	if err := os.MkdirAll(fbScheduleOutDir, 0755); err != nil {
		logFatal("Failed to create output folder %s: %v", fbScheduleOutDir, err)
	}

	// Get current working directory for video path
	cwd, err := os.Getwd()
	if err != nil {
		logFatal("Failed to get current working directory: %v", err)
	}

	var posts []FBScheduleEntry

	for _, entry := range payload.Prompts {
		if entry.Tag == "" {
			continue
		}

		mdPath := filepath.Join(geminiOutDir, entry.Tag+".md")
		mdBytes, err := os.ReadFile(mdPath)
		if err != nil {
			logFatal("Failed to read %s for tag '%s': %v", mdPath, entry.Tag, err)
		}

		mdText := string(mdBytes)

		// Extract CAPTION content (between "CAPTION:" and "## MAIN PROMPT" or end of CAPTION section)
		caption := extractCaption(mdText)
		if caption == "" {
			logFatal("CAPTION not found in %s for tag '%s'", mdPath, entry.Tag)
		}

		// Build video path: {{current_path}}/.data-workflow-googleflow-out/{{tag}}.mp4
		videoPath := filepath.Join(cwd, googleflowOutDir, entry.Tag+".mp4")
		// Convert to forward slashes for JSON
		videoPath = filepath.ToSlash(videoPath)

		posts = append(posts, FBScheduleEntry{
			Caption:   caption,
			VideoPath: videoPath,
			SchedulingInfo: SchedulingInfo{
				IsSuccess: false,
				Tried:     false,
			},
		})
		fmt.Printf("Processed tag '%s' -> caption: %d chars\n", entry.Tag, len(caption))
	}

	fbPayload := FBSchedulePayload{
		FbPageSlug: *fbPageSlug,
		PageName:   *pageName,
		Posts:      posts,
	}

	outBytes, err := json.MarshalIndent(fbPayload, "", "    ")
	if err != nil {
		logFatal("Failed to marshal fbSchedule metadata: %v", err)
	}
	outBytes = append(outBytes, '\n')

	outPath := filepath.Join(fbScheduleOutDir, "metadata.json")
	if err := os.WriteFile(outPath, outBytes, 0644); err != nil {
		logFatal("Failed to write %s: %v", outPath, err)
	}
	fmt.Printf("Wrote %s with %d posts\n", outPath, len(posts))
}

// extractCaption extracts the caption text from the markdown file.
// The caption is between "CAPTION:" and "## MAIN PROMPT (copy-paste)" or the next major section.
func extractCaption(mdText string) string {
	captionMarker := "CAPTION:"
	startIdx := strings.Index(mdText, captionMarker)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(captionMarker)

	// Find the end - either "## MAIN PROMPT" or "## " (next section)
	endMarkers := []string{"## MAIN PROMPT", "\n## ", "\n# "}
	endIdx := -1
	for _, marker := range endMarkers {
		idx := strings.Index(mdText[startIdx:], marker)
		if idx != -1 {
			if endIdx == -1 || idx < endIdx {
				endIdx = idx
			}
		}
	}

	if endIdx == -1 {
		// No end marker found, take rest of file
		return strings.TrimSpace(mdText[startIdx:])
	}

	caption := mdText[startIdx : startIdx+endIdx]
	return strings.TrimSpace(caption)
}

func runClearTmpData() {
	wd, err := os.Getwd()
	if err != nil {
		logFatal("Failed to get current working directory: %v", err)
	}

	entries, err := os.ReadDir(wd)
	if err != nil {
		logFatal("Failed to read directory %s: %v", wd, err)
	}

	deleted := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".data") {
			continue
		}
		path := filepath.Join(wd, name)
		if err := os.RemoveAll(path); err != nil {
			logFatal("Failed to remove %s: %v", path, err)
		}
		fmt.Printf("Removed %s\n", path)
		deleted++
	}
	fmt.Printf("Cleared %d .data* entries in %s\n", deleted, wd)
}

func runNewWorkflowFBScheduling() {
	geminiInDir := ".data-harness-gemini-in"

	if err := os.MkdirAll(geminiInDir, 0755); err != nil {
		logFatal("Failed to create input folder %s: %v", geminiInDir, err)
	}

	exampleEntry := PromptEntry{
		Tag:    "example-tag",
		Images: []string{"./example-image.png"},
		Text:   "Create an ultra-length prompt for this subject. A detailed scene description here.",
	}
	examplePayload := InputPayload{
		Prompts: []PromptEntry{exampleEntry},
	}

	outBytes, err := json.MarshalIndent(examplePayload, "", "    ")
	if err != nil {
		logFatal("Failed to marshal example human-prompts.json: %v", err)
	}
	outBytes = append(outBytes, '\n')

	promptsPath := filepath.Join(geminiInDir, "human-prompts.json")
	if err := os.WriteFile(promptsPath, outBytes, 0644); err != nil {
		logFatal("Failed to write %s: %v", promptsPath, err)
	}
	fmt.Printf("Created %s with example prompt\n", promptsPath)
	fmt.Println("Edit the file to add your tags, images, and prompt texts.")
	fmt.Println("Then run the Gemini harness to generate .md files in .data-harness-gemini-out/")
}

func logFatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
