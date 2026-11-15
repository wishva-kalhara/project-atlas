package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type PromptEntry struct {
	Tag    string   `json:"tag"`
	Images []string `json:"images"`
	Text   string   `json:"text"`
}

type InputPayload struct {
	Prompts []PromptEntry `json:"prompts"`
}

type evalStatus struct {
	Count      int    `json:"count"`
	Generating bool   `json:"generating"`
	LatestText string `json:"latestText"`
}

func main() {
	chatURL := flag.String("url", "", "Full URL of your existing Gemini chat (e.g. https://gemini.google.com/app/abc123xyz)")
	inputDir := flag.String("input", ".input", "Folder containing human-prompts.json and images")
	outputDir := flag.String("output", ".data", "Folder where <tag>.md files will be saved")
	profileDir := flag.String("profile", ".chrome-profile", "Folder to persist your Google login session")
	flag.Parse()

	if *chatURL == "" {
		log.Fatal("Error: Provide your existing chat URL using -url (e.g. harness.exe -url \"https://gemini.google.com/app/...\")")
	}

	// 1. Read human-prompts.json
	promptsPath := filepath.Join(*inputDir, "human-prompts.json")
	rawJSON, err := os.ReadFile(promptsPath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", promptsPath, err)
	}

	var payload InputPayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		log.Fatalf("Invalid JSON in %s: %v", promptsPath, err)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output folder %s: %v", *outputDir, err)
	}

	absProfile, err := filepath.Abs(*profileDir)
	if err != nil {
		log.Fatalf("Failed to resolve profile path: %v", err)
	}

	// 2. Launch visible Chrome with a persistent profile so your Google login is saved
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserDataDir(absProfile),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	// 3. Navigate directly to your existing Gemini conversation
	fmt.Printf("Opening chat: %s\n", *chatURL)
	if err := chromedp.Do(ctx,
		chromedp.Navigate(*chatURL),
		chromedp.Sleep(4*time.Second),
	); err != nil {
		log.Fatalf("Failed to open chat URL: %v", err)
	}

	// Check if login is required on first run
	currentURL, _ := chromedp.Run(ctx, chromedp.Location())
	if strings.Contains(currentURL, "accounts.google.com") {
		fmt.Println("Please sign in to your Google account in the opened Chrome window.")
		fmt.Println("Press ENTER here once your Gemini chat is loaded...")
		fmt.Scanln()
	}

	// 4. Process each prompt strictly serially in the same chat
	for i, entry := range payload.Prompts {
		if entry.Tag == "" {
			continue
		}
		fmt.Printf("[%d/%d] Processing '%s'...\n", i+1, len(payload.Prompts), entry.Tag)

		// Count existing responses before sending so we know when the new one arrives
		prevCount, _ := chromedp.Run(ctx, chromedp.Evaluate[int](`document.querySelectorAll('model-response').length`))

		// Resolve all image paths for this prompt
		var absImages []string
		for _, imgRef := range entry.Images {
			imgPath, err := resolveImagePath(*inputDir, imgRef)
			if err != nil {
				log.Fatalf("Image error on '%s': %v", entry.Tag, err)
			}
			absPath, _ := filepath.Abs(imgPath)
			absImages = append(absImages, absPath)
			fmt.Printf("  + Queued image: %s\n", filepath.Base(absPath))
		}

		// Upload images if present
		if len(absImages) > 0 {
			if err := uploadImages(ctx, absImages); err != nil {
				log.Fatalf("Failed to upload images for '%s': %v", entry.Tag, err)
			}
		}

		// Type prompt into the rich-textarea input box and submit
		if err := sendPromptText(ctx, entry.Text); err != nil {
			log.Fatalf("Failed to send prompt '%s': %v", entry.Tag, err)
		}

		// Wait until Gemini finishes generating the response
		fmt.Println("  ... Waiting for Gemini response")
		responseText, err := waitForResponse(ctx, prevCount, 5*time.Minute)
		if err != nil {
			log.Fatalf("Timeout or error waiting for response on '%s': %v", entry.Tag, err)
		}

		// Save output to .data/<tag>.md
		outPath := filepath.Join(*outputDir, entry.Tag+".md")
		if err := os.WriteFile(outPath, []byte(responseText), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", outPath, err)
		}
		fmt.Printf("  -> Saved %s\n", outPath)

		// Brief pause before next serial turn
		time.Sleep(2 * time.Second)
	}

	fmt.Println("All prompts completed.")
}

func uploadImages(ctx context.Context, absPaths []string) error {
	hasInput, _ := chromedp.Run(ctx, chromedp.Evaluate[bool](`document.querySelector('input[type="file"]') !== null`))
	if !hasInput {
		_ = chromedp.Do(ctx,
			chromedp.Click(`button[aria-label*="Upload"], button[aria-label*="Add files"], button.upload-card-button`),
			chromedp.Sleep(1*time.Second),
		)
	}

	return chromedp.Do(ctx,
		chromedp.SetUploadFiles(`input[type="file"]`, absPaths),
		chromedp.Sleep(3*time.Second),
	)
}

func sendPromptText(ctx context.Context, text string) error {
	selector := `div.ql-editor[contenteditable="true"], rich-textarea div[contenteditable="true"]`
	return chromedp.Do(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Click(selector),
		chromedp.SendKeys(selector, text),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`button[aria-label*="Send"], button.send-button`),
	)
}

func waitForResponse(ctx context.Context, prevCount int, timeout time.Duration) (string, error) {
	js := `(() => {
		const responses = document.querySelectorAll('model-response');
		const stopBtn = document.querySelector('button[aria-label*="Stop"]');
		const latest = responses.length > 0 ? responses[responses.length - 1] : null;
		const content = latest ? (latest.querySelector('.markdown, message-content') || latest) : null;
		return {
			count: responses.length,
			generating: stopBtn !== null,
			latestText: content ? content.innerText : ""
		};
	})()`

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := chromedp.Run(ctx, chromedp.Evaluate[evalStatus](js))
		if err != nil {
			return "", err
		}

		// New response appeared AND the "Stop generating" button is gone
		if status.Count > prevCount && !status.Generating && strings.TrimSpace(status.LatestText) != "" {
			time.Sleep(1 * time.Second)
			finalStatus, err := chromedp.Run(ctx, chromedp.Evaluate[evalStatus](js))
			if err == nil && strings.TrimSpace(finalStatus.LatestText) != "" {
				return finalStatus.LatestText, nil
			}
			return status.LatestText, nil
		}

		time.Sleep(1 * time.Second)
	}
	return "", fmt.Errorf("timed out after %v", timeout)
}

func resolveImagePath(inputDir, ref string) (string, error) {
	candidate := filepath.Join(inputDir, filepath.Clean(ref))
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate, nil
	}
	if matches, err := filepath.Glob(candidate + ".*"); err == nil && len(matches) > 0 {
		return matches[0], nil
	}
	return "", fmt.Errorf("could not find image '%s' in '%s'", ref, inputDir)
}
