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

	// Load metadata from the googleflow-in directory, resolved relative to the
	// executable's current working directory (not the workspace root).
	logger.Info("GoogleFlowWorkflow:Run", "Loading metadata from .data-workflow-googleflow-in/metadata.json")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	metadataPath := filepath.Join(cwd, ".data-workflow-googleflow-in", "metadata.json")
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	var metadataList []GoogleFlowMetadata
	if err := json.Unmarshal(metadataBytes, &metadataList); err != nil {
		return fmt.Errorf("failed to parse metadata JSON: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Loaded %d metadata entries", len(metadataList)))

	if index < 0 || index >= len(metadataList) {
		return fmt.Errorf("index %d out of range (0-%d)", index, len(metadataList)-1)
	}

	entry := metadataList[index]
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Processing entry index=%d tag=%s", index, entry.Tag))

	if len(entry.Images) == 0 {
		return fmt.Errorf("no images found for entry at index %d", index)
	}

	logger.Info("GoogleFlowWorkflow:Run", "Initializing browser session")
	session, err := browser.NewSession(g.cfg.CookiesPath, headed)
	if err != nil {
		return fmt.Errorf("failed to init browser session: %w", err)
	}
	defer session.Close()

	logger.Info("GoogleFlowWorkflow:Run", "Creating new page")
	page, err := session.Context.NewPage()
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Navigating to Google Flow URL: %s", g.cfg.GoogleFlowURL))
	if _, err := page.Goto(g.cfg.GoogleFlowURL); err != nil {
		return fmt.Errorf("failed to go to Google Flow URL: %w", err)
	}

	// Click "Create with Google Flow" button if present
	logger.Info("GoogleFlowWorkflow:Run", "Checking for 'Create with Google Flow' button")
	createBtn := page.GetByLabel("Create with Google Flow")
	if visible, _ := createBtn.IsVisible(); visible {
		logger.Info("GoogleFlowWorkflow:Run", "Clicking 'Create with Google Flow' button")
		_ = createBtn.Click()
		logger.Info("GoogleFlowWorkflow:Run", "Clicked 'Create with Google Flow' button")
	} else {
		logger.Info("GoogleFlowWorkflow:Run", "'Create with Google Flow' button not visible, skipping")
	}

	logger.Info("GoogleFlowWorkflow:Run", "Waiting 5 seconds for page to settle")
	time.Sleep(5 * time.Second)

	// Handle sign-in redirect
	current := page.URL()
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Current URL after navigation: %s", current))
	if !(current == "" || current == "https://flow.google.com/" || (len(current) > 25 && current[:25] == "https://accounts.google.com")) {
		logger.Info("GoogleFlowWorkflow:Run", "No redirect needed, continuing")
	} else {
		logger.Info("GoogleFlowWorkflow:Run", "Redirected to Google sign-in. Waiting for manual sign-in (2 min)...")
		if err := page.WaitForURL("https://flow.google.com/**", playwright.PageWaitForURLOptions{Timeout: playwright.Float(120000)}); err != nil {
			return fmt.Errorf("timeout waiting for Google login: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", "Login completed")
	}

	logger.Info("GoogleFlowWorkflow:Run", "Waiting for flow.google.com URL")
	page.WaitForURL("https://flow.google.com/")
	logger.Info("GoogleFlowWorkflow:Run", "URL resolved to flow.google.com")

	// Click New Project
	logger.Info("GoogleFlowWorkflow:Run", "Looking for 'New Project' button")
	newProj := page.Locator("text=New Project")
	if err := newProj.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err == nil {
		logger.Info("GoogleFlowWorkflow:Run", "Clicking 'New Project' button")
		_ = newProj.Click()
		logger.Info("GoogleFlowWorkflow:Run", "Clicked 'New Project' button, waiting 5 seconds")
		page.WaitForTimeout(5000)
	} else {
		logger.Info("GoogleFlowWorkflow:Run", "'New Project' button not found or not visible")
	}

	// Upload images one by one
	for i, imagePath := range entry.Images {
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Uploading image %d/%d: %s", i+1, len(entry.Images), imagePath))

		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicking 'Add ingredients to the prompt' for image %d", i+1))
		if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Add ingredients to the prompt",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Add ingredients to the prompt: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicked 'Add ingredients to the prompt' for image %d", i+1))

		// Intercept the file chooser before clicking "Upload media" so the native OS dialog never opens
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Opening file chooser via 'Upload media' for image %d", i+1))
		fileChooser, err := page.ExpectFileChooser(func() error {
			return page.GetByRole("button", playwright.PageGetByRoleOptions{
				Name: "Upload media",
			}).Click()
		})
		if err != nil {
			return fmt.Errorf("failed to open file chooser via Upload media: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  File chooser opened for image %d", i+1))

		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Setting file: %s", imagePath))
		if err := fileChooser.SetFiles(imagePath); err != nil {
			return fmt.Errorf("failed to set file %s: %w", imagePath, err)
		}
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  File set for image %d", i+1))

		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicking 'Loading preview' progressbar for image %d", i+1))
		if err := page.GetByRole("progressbar", playwright.PageGetByRoleOptions{
			Name: "Loading preview",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Loading preview progressbar: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicked 'Loading preview' progressbar for image %d", i+1))

		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Waiting 5 seconds for preview to load (image %d)", i+1))
		page.WaitForTimeout(5000)

		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicking 'Add to prompt' for image %d", i+1))
		if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Add to prompt",
		}).Click(); err != nil {
			return fmt.Errorf("failed to click Add to prompt: %w", err)
		}
		logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("  Clicked 'Add to prompt' for image %d", i+1))
	}
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("All %d images uploaded successfully", len(entry.Images)))

	// Enter the video prompt
	logger.Info("GoogleFlowWorkflow:Run", "Entering video prompt")
	if err := page.GetByRole("textbox", playwright.PageGetByRoleOptions{
		Name: "What do you want to create?",
	}).GetByRole("paragraph").Click(); err != nil {
		return fmt.Errorf("failed to click prompt paragraph: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Clicked prompt paragraph")

	// Fill the prompt
	if err := page.GetByRole("textbox", playwright.PageGetByRoleOptions{
		Name: "What do you want to create?",
	}).GetByRole("paragraph").Fill(entry.VideoPrompt); err != nil {
		return fmt.Errorf("failed to fill prompt: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Filled video prompt (%d chars)", len(entry.VideoPrompt)))

	// Select video config
	logger.Info("GoogleFlowWorkflow:Run", "Opening video settings")
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Settings trigger",
	}).Click(); err != nil {
		return fmt.Errorf("failed to click Settings trigger: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Clicked Settings trigger")

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: "Video",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select Video radio: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Selected 'Video' radio")

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: ":16",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select :16 radio: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Selected ':16' radio")

	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Select model family",
	}).Click(); err != nil {
		return fmt.Errorf("failed to click Select model family: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Clicked 'Select model family'")

	if err := page.GetByRole("menuitem", playwright.PageGetByRoleOptions{
		Name: "Omni 1.1 Flash Leaving Dec",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select model menuitem: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Selected 'Omni 1.1 Flash Leaving Dec' model")

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: "10s",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select 10s radio: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Selected '10s' radio")

	if err := page.GetByRole("radio", playwright.PageGetByRoleOptions{
		Name: "x1",
	}).Click(); err != nil {
		return fmt.Errorf("failed to select x1 radio: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Selected 'x1' radio")

	logger.Info("GoogleFlowWorkflow:Run", "Waiting 5 seconds before starting generation")
	page.WaitForTimeout(5_000)

	logger.Info("GoogleFlowWorkflow:Run", "Clicking 'Start generation'")
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Start generation",
	}).Click(); err != nil {
		return fmt.Errorf("failed to click Start generation: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Clicked 'Start generation', waiting for video to be generated")

	logger.Info("GoogleFlowWorkflow:Run", "Waiting for generated video thumbnail to appear")
	// Wait for the generated video thumbnail to become visible
	thumbnail := page.GetByRole("img", playwright.PageGetByRoleOptions{
		Name: "Generated video thumbnail",
	})
	if err := thumbnail.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(600000)}); err != nil {
		return fmt.Errorf("timeout waiting for generated video thumbnail: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Generated video thumbnail is visible")

	page.WaitForTimeout(5_000)

	// Select generated video thumbnail
	logger.Info("GoogleFlowWorkflow:Run", "Clicking generated video thumbnail")
	err = page.GetByRole("img", playwright.PageGetByRoleOptions{
		Name: "Generated video thumbnail",
	}).Click()
	if err != nil {
		return fmt.Errorf("failed to click Generated video thumbnail: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Clicked generated video thumbnail")

	// Open download menu
	logger.Info("GoogleFlowWorkflow:Run", "Opening download menu")
	err = page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Download media",
	}).Click()
	if err != nil {
		return fmt.Errorf("failed to click Download media: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", "Opened download menu")

	// Resolve current working directory and create output folder
	logger.Info("GoogleFlowWorkflow:Run", "Preparing output directory")
	cwd, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	outDir := filepath.Join(cwd, ".data-workflow-googleflow-out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Output directory ready: %s", outDir))

	// Use tag-based filename
	outputPath := filepath.Join(outDir, entry.Tag+".mp4")
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Target output path: %s", outputPath))

	// Trigger download and save to target path
	logger.Info("GoogleFlowWorkflow:Run", "Triggering download (720p Original size)")
	download, err := page.ExpectDownload(func() error {
		return page.GetByRole("menuitem", playwright.PageGetByRoleOptions{
			Name: "720p Original size",
		}).Click()
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Download started: %s", download.SuggestedFilename()))

	if err := download.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save download to %s: %w", outputPath, err)
	}

	logger.Info("GoogleFlowWorkflow:Run", fmt.Sprintf("Downloaded (%s) and saved to %s", download.SuggestedFilename(), outputPath))

	logger.Info("GoogleFlowWorkflow:Run", "Google Flow Video Gen Workflow completed successfully")
	return nil
}
