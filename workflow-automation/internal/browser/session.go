package browser

import (
	"fmt"
	"os"

	"workflow-automation/internal/logger"

	"github.com/mxschmitt/playwright-go"
)

type Session struct {
	PW         *playwright.Playwright
	Browser    playwright.Browser
	Context    playwright.BrowserContext
	CookiePath string
}

// NewSession starts Playwright, launches Chrome (or fallback) and creates a browser context.
func NewSession(cookiePath string, headed bool) (*Session, error) {
	logger.Debug("Browser:NewSession", "Starting Playwright driver")

	runOptions := &playwright.RunOptions{}

	pw, err := playwright.Run(runOptions)
	if err != nil {
		return nil, fmt.Errorf("could not launch Playwright: %w", err)
	}

	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(!headed),
		Channel:  playwright.String("chrome"), // Prefer system Chrome
		Args: []string{
			"--disable-blink-features=AutomationControlled",
		},
		IgnoreDefaultArgs: []string{"--enable-automation"},
	}

	browser, err := pw.Chromium.Launch(launchOptions)
	if err != nil {
		// Fallback to bundled Chromium if Chrome channel fails
		logger.Warn("Browser:NewSession", "Host Chrome channel failed; falling back to bundled Chromium")
		launchOptions.Channel = nil
		browser, err = pw.Chromium.Launch(launchOptions)
		if err != nil {
			_ = pw.Stop()
			return nil, fmt.Errorf("could not launch browser: %w", err)
		}
	}

	contextOptions := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 720},
	}
	// Load stored cookies if present
	if _, err := os.Stat(cookiePath); err == nil {
		logger.Info("Browser:NewSession", fmt.Sprintf("Loading storage state from %s", cookiePath))
		contextOptions.StorageStatePath = playwright.String(cookiePath)
	} else {
		logger.Info("Browser:NewSession", "No cookies file found; starting clean session")
	}

	context, err := browser.NewContext(contextOptions)
	if err != nil {
		_ = browser.Close()
		_ = pw.Stop()
		return nil, fmt.Errorf("could not create browser context: %w", err)
	}

	return &Session{PW: pw, Browser: browser, Context: context, CookiePath: cookiePath}, nil
}

// SaveCookies persists the current storage state to the configured file.
func (s *Session) SaveCookies() error {
	logger.Debug("Browser:SaveCookies", fmt.Sprintf("Saving storage state to %s", s.CookiePath))
	if _, err := s.Context.StorageState(playwright.BrowserContextStorageStateOptions{Path: playwright.String(s.CookiePath)}); err != nil {
		return fmt.Errorf("failed to save storage state: %w", err)
	}
	logger.Info("Browser:SaveCookies", "Storage state saved successfully")
	return nil
}

// Close cleans up the session, persisting cookies first.
func (s *Session) Close() {
	if s.Context != nil {
		_ = s.SaveCookies()
		_ = s.Context.Close()
	}
	if s.Browser != nil {
		_ = s.Browser.Close()
	}
	if s.PW != nil {
		_ = s.PW.Stop()
	}
}
