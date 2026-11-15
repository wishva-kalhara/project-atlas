package pages

import (
    "fmt"
    "strings"

    "workflow-automation/internal/logger"
    "github.com/mxschmitt/playwright-go"
)

type FbSchedulingPage struct {
    BasePage
    baseURL               string
    scheduledContentURL   string
}

func NewFbSchedulingPage(page playwright.Page, baseURL, scheduledURL string) *FbSchedulingPage {
    return &FbSchedulingPage{
        BasePage:            NewBasePage(page),
        baseURL:             baseURL,
        scheduledContentURL: scheduledURL,
    }
}

func (f *FbSchedulingPage) Open(pageSlug string) error {
    targetURL := fmt.Sprintf("%s%s/", strings.TrimRight(f.baseURL, "/")+"/", pageSlug)
    logger.Debug("FbSchedulingPage:Open", "Navigating to "+targetURL)
    _, err := f.Page.Goto(targetURL)
    return err
}

func (f *FbSchedulingPage) HandleLoginIfRequired(pageSlug, pageName string) error {
    logger.Debug("FbSchedulingPage:HandleLoginIfRequired", "Checking for login indicator")
    loginIndicator := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Log in to Facebook"}).First()
    visible, err := loginIndicator.IsVisible()
    if err == nil && visible {
        logger.Info("FbSchedulingPage:HandleLoginIfRequired", "Login dialog detected. Waiting up to 3 minutes for manual sign-in...")
        expectedURL := fmt.Sprintf("%s/%s", strings.TrimRight(f.baseURL, "/"), pageSlug)
        err = f.Page.WaitForURL(expectedURL, playwright.PageWaitForURLOptions{Timeout: playwright.Float(180000)})
        if err != nil {
            return fmt.Errorf("timeout waiting for login completion: %w", err)
        }
        // Wait for password field to detach
        _ = f.Page.Locator("input[type=\"password\"]").WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden, Timeout: playwright.Float(180000)})
        logger.Info("FbSchedulingPage:HandleLoginIfRequired", "Manual sign-in completed")
        // Switch profile if necessary
        accBtn := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Account, your profile"})
        if err := accBtn.Click(); err == nil {
            switchBtn := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: fmt.Sprintf("Switch to %s", pageName)})
            _ = switchBtn.Click()
            logger.Info("FbSchedulingPage:HandleLoginIfRequired", "Switched to "+pageName)
        }
    } else {
        logger.Info("FbSchedulingPage:HandleLoginIfRequired", "Session already authenticated")
    }
    return nil
}

func (f *FbSchedulingPage) WaitForExactPageURL(pageSlug string) error {
    expectedURL := fmt.Sprintf("%s/%s", strings.TrimRight(f.baseURL, "/"), pageSlug)
    logger.Debug("FbSchedulingPage:WaitForExactPageURL", "Awaiting "+expectedURL)
    return f.Page.WaitForURL(expectedURL, playwright.PageWaitForURLOptions{Timeout: playwright.Float(120000)})
}

func (f *FbSchedulingPage) NavigateToScheduledContent() error {
    logger.Debug("FbSchedulingPage:NavigateToScheduledContent", "Navigating to Content Library")
    // Option A: sidebar navigation
    creatorLink := f.Page.GetByRole("main", playwright.PageGetByRoleOptions{Name: "Page profile content"}).GetByRole("link", playwright.LocatorGetByRoleOptions{Name: "Creator Dashboard"})
    if err := creatorLink.Click(); err == nil {
        _ = f.Page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Content", Exact: playwright.Bool(true)}).Click()
        _ = f.Page.GetByRole("tab", playwright.PageGetByRoleOptions{Name: "Scheduled"}).Click()
        return nil
    }
    // Option B: direct URL fallback
    _, err := f.Page.Goto(f.scheduledContentURL)
    return err
}

func (f *FbSchedulingPage) ClickCreateButton() error {
    logger.Debug("FbSchedulingPage:ClickCreateButton", "Clicking Create button")
    createBtn := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Create", Exact: playwright.Bool(true)})
    if err := createBtn.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err == nil {
        return createBtn.Click()
    }
    // Fallback to XPath selector
    return f.Page.Click("xpath=//*[@id=\"prodash-create-button\"]/div")
}

func (f *FbSchedulingPage) UploadReelVideo(videoAbsPath string) error {
    logger.Debug("FbSchedulingPage:UploadReelVideo", "Uploading video: "+videoAbsPath)
    reelMenuItem := f.Page.GetByRole("menuitem", playwright.PageGetByRoleOptions{Name: "Reel", Exact: playwright.Bool(true)})
    fileChooser, err := f.Page.ExpectFileChooser(func() error {
        if err := reelMenuItem.Click(); err != nil {
            // Fallback click path
            return f.Page.Click("xpath=//*[@id=\"prodash_create_menu_items\"]/div/div/div/div/div[1]/div/div[2]")
        }
        return nil
    })
    if err != nil {
        return fmt.Errorf("failed to trigger file chooser: %w", err)
    }
    if err := fileChooser.SetFiles(videoAbsPath); err != nil {
        return fmt.Errorf("failed to supply file to chooser: %w", err)
    }
    logger.Info("FbSchedulingPage:UploadReelVideo", "Video attached successfully")
    return nil
}

func (f *FbSchedulingPage) FillDetailsAndSchedule(caption string) error {
    logger.Debug("FbSchedulingPage:FillDetailsAndSchedule", "Filling caption and scheduling")
    // Wait for video preview ingestion
    f.Page.WaitForTimeout(5000)
    // Focus editor and fill caption
    _ = f.Page.GetByRole("paragraph").Click()
    descBox := f.Page.GetByRole("textbox", playwright.PageGetByRoleOptions{Name: "Describe your reel..."})
    if err := descBox.Fill(caption); err != nil {
        return fmt.Errorf("could not fill caption: %w", err)
    }
    // Click Next
    if err := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Next"}).Click(); err != nil {
        return fmt.Errorf("could not click Next: %w", err)
    }
    // Click Schedule
    if err := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Schedule"}).Click(); err != nil {
        return fmt.Errorf("could not click Schedule: %w", err)
    }
    // Wait for scheduling dialog to disappear (up to 10 min)
    dialog := f.Page.GetByRole("dialog", playwright.PageGetByRoleOptions{Name: "Scheduling options"}).Locator("div").Filter(playwright.LocatorFilterOptions{HasText: "^Scheduling options$"})
    _ = dialog.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden, Timeout: playwright.Float(600000)})
    // Click final Schedule post button
    schedulePostBtn := f.Page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Schedule post"})
    if err := schedulePostBtn.Click(); err != nil {
        return fmt.Errorf("could not click Schedule post: %w", err)
    }
    logger.Info("FbSchedulingPage:FillDetailsAndSchedule", "Reel scheduled successfully on Facebook")
    return nil
}
