package workflows

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"workflow-automation/internal/browser"
	"workflow-automation/internal/config"
	"workflow-automation/internal/logger"

	"github.com/mxschmitt/playwright-go"
)

// GoogleFlowMetadata represents a single entry in the metadata.json
type GoogleFlowMetadata struct {
	Tag         string   `json:"tag"`
	Images      []string `json:"images"`
	VideoPrompt string   `json:"videoPrompt"`
}

type GoogleFlowWorkflow struct {
	cfg *config.Config
}

func NewGoogleFlowWorkflow(cfg *config.Config) *GoogleFlowWorkflow {
	return &GoogleFlowWorkflow{cfg: cfg}
}

func (g *GoogleFlowWorkflow) Run(headed bool, index int) error {
	logger.Info("GoogleFlowWorkflow:Run", "Starting Google Flow Video Gen Workflow")

	// Load metadata from the googleflow-in directory
	metadataPath := filepath.Join(g.cfg.WorkspaceRoot, ".data-workflow-googleflow-in", "metadata.json")
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	var metadataList []GoogleFlowMetadata
	if err := json.Unmarshal(metadataBytes, &metadataList); err != nil {
		return fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	if index < 0 || index >= len(metadataList) {
		return fmt.Errorf("index %d out of range (0-%d)", index, len(metadataList)-1)
	}

	entry := metadataList[index]
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Processing entry index=%d tag=%s", index, entry.Tag))

	if len(entry.Images) == 0 {
		return fmt.Errorf("no images found for entry at index %d", index)
	}

	session, err := browser.NewSession(g.cfg.CookiesPath, headed)
	if err != nil {
		return fmt.Errorf("failed to init browser session: %w", err)
	}
	defer session.Close()

	page, err := session.Context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	logger.Debug("GoogleFlowWorkflow:Run", "Opening "+g.cfg.GoogleFlowURL)
	if _, err := page.Goto(g.cfg.GoogleFlowURL); err != nil {
		return fmt.Errorf("failed to go to Google Flow URL: %w", err)
	}

	// Click "Create with Google Flow" button if present
	createBtn := page.GetByLabel("Create with Google Flow")
	if visible, _ := createBtn.IsVisible(); visible {
		_ = createBtn.Click()
	}

	time.Sleep(5 * time.Second)

	// Handle sign-in redirect
	current := page.URL()
	if !(current == "" || current == "https://flow.google.com/" || (len(current) > 25 && current[:25] == "https://accounts.google.com")) {
		// No redirect, continue
	} else {
		logger.Info("GoogleFlowWorkflow:Run", "Redirected to Google sign-in. Waiting for manual sign-in (2 min)...")
		if err := page.WaitForURL("https://flow.google.com/**", playwright.PageWaitForURLOptions{Timeout: playwright.Float(120000)}); err != nil {
			return fmt.Errorf("timeout waiting for Google login: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", "Login completed")
	}

	page.WaitForURL("https://flow.google.com/")

	// Click New Project
	newProj := page.Locator("text=New Project")
	if err := newProj.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err == nil {
		_ = newProj.Click()
		page.WaitForTimeout(5000)
	}

	// Upload images one by one
	for i, imagePath := range entry.Images {
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Uploading image %d/%d: %s", i+1, len(entry.Images), imagePath))

		if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Add ingredients to the prompt",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Add ingredients to the prompt: %w", err)
		}

		// Intercept the file chooser before clicking "Upload media" so the native OS dialog never opens
		fileChooser, err := page.ExpectFileChooser(func() error {
			return page.GetByRole("button", playwright.PageGetByRoleOptions{
				Name: "Upload media",
			}).Click()
		})
		if err != nil {
			return fmt.Errorf("failed to open file chooser via Upload media: %w", err)
		}

		if err := fileChooser.SetFiles(imagePath); err != nil {
			return fmt.Errorf("failed to set file %s: %w", imagePath, err)
		}

		if err := page.GetByRole("progressbar", playwright.PageGetByRoleOptions{
			Name: "Loading preview",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Loading preview progressbar: %w", err)
		}

		page.WaitForTimeout(5000)

		if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Add to prompt",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Add to prompt: %w", err)
		}
	}

	// Enter the video prompt
	if err := page.GetByRole("textbox", playwright.PageGetByRoleOptions{
		Name: "What do you want to create?",
	}).GetByRole("paragraph").Click(); err != nil {
		return fmt.Errorf("failed to click prompt paragraph: %w", err)
	}

	// Fill the prompt
	if err := page.GetByRole("textbox", playwright.PageGetByRoleOptions{
		Name: "What do you want to create?",
	}).GetByRole("paragraph").Fill(entry.VideoPrompt); err != nil {
		return fmt.Errorf("failed to fill prompt: %w", err)
	}

	// Select video config
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Settings trigger",
	}).Click(); err != nil {
		return fmt.Errorf("failed to click Settings trigger: %w", err)
	}

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: "Video",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select Video radio: %w", err)
	}

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: ":16",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select :16 radio: %w", err)
	}

	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Select model family",
	}).Click(); err != nil {
		return fmt.Errorf("failed to click Select model family: %w", err)
	}

	if err := page.GetByRole("menuitem", playwright.PageGetByRoleOptions{
		Name: "Omni 1.1 Flash Leaving Dec",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select model menuitem: %w", err)
	}

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: "10s",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select 10s radio: %w", err)
	}

	page.Pause()
	page.WaitForTimeout(5_000)

	// Select generated video thumbnail
	err = page.GetByRole("img", playwright.PageGetByRoleOptions{
		Name: "Generated video thumbnail",
	}).Click()
	if err != nil {
		return fmt.Errorf("failed to click Generated video thumbnail: %w", err)
	}

	// Open download menu
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Download media",
	}).Click()
	if err != nil {
		return fmt.Errorf("failed to click Download media: %w", err)
	}

	// Resolve current working directory and create output folder
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	outDir := filepath.Join(cwd, ".data-workflow-googleflow-out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use tag-based filename
	outputPath := filepath.Join(outDir, entry.Tag+".mp4")

	// Trigger download and save to target path
	download, err := page.ExpectDownload(func() error {
		return page.GetByRole("menuitem", playwright.PageGetByRoleOptions{
			Name: "720p Original size",
		}).Click()
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := download.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save download to %s: %w", outputPath, err)
	}

	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Downloaded (%s) and saved to %s", download.SuggestedFilename(), outputPath))

	return nil
}
