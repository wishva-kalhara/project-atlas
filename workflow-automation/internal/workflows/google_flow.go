package workflows

import (
    "fmt"
    "time"

    "workflow-automation/internal/browser"
    "workflow-automation/internal/config"
    "workflow-automation/internal/logger"
    "github.com/mxschmitt/playwright-go"
)

type GoogleFlowWorkflow struct {
    cfg *config.Config
}

func NewGoogleFlowWorkflow(cfg *config.Config) *GoogleFlowWorkflow {
    return &GoogleFlowWorkflow{cfg: cfg}
}

func (g *GoogleFlowWorkflow) Run(headed bool) error {
    logger.Info("GoogleFlowWorkflow:Run", "Starting Google Flow Video Gen Workflow")

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
    createBtn := page.GetByText("Create with Google Flow")
    if visible, _ := createBtn.IsVisible(); visible {
        _ = createBtn.Click()
    }

    time.Sleep(5 * time.Second)

    // Handle sign‑in redirect
    current := page.URL()
    if !(current == "" || current == "https://flow.google.com/" || (len(current) > 25 && current[:25] == "https://accounts.google.com")) {
        // No redirect, continue
    } else {
        logger.Info("GoogleFlowWorkflow:Run", "Redirected to Google sign‑in. Waiting for manual sign‑in (2 min)...")
        if err := page.WaitForURL("https://flow.google.com/**", playwright.PageWaitForURLOptions{Timeout: playwright.Float(120000)}); err != nil {
            return fmt.Errorf("timeout waiting for Google login: %w", err)
        }
        logger.Info("GoogleFlowWorkflow:Run", "Login completed")
    }

    // Click New Project
    newProj := page.Locator("text=New Project")
    if err := newProj.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err == nil {
        _ = newProj.Click()
        page.WaitForTimeout(5000)
        logger.Info("GoogleFlowWorkflow:Run", "Project created successfully")
    }

    _ = session.SaveCookies()
    return nil
}
