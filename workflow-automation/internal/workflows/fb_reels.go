package workflows

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "workflow-automation/internal/browser"
    "workflow-automation/internal/config"
    "workflow-automation/internal/data"
    "workflow-automation/internal/logger"
    "workflow-automation/internal/pages"
)

type FbReelsWorkflow struct {
    cfg  *config.Config
    repo *data.Repository
}

func NewFbReelsWorkflow(cfg *config.Config, repo *data.Repository) *FbReelsWorkflow {
    return &FbReelsWorkflow{cfg: cfg, repo: repo}
}

func (w *FbReelsWorkflow) Run(headed bool) error {
    logger.Info("FbReelsWorkflow:Run", "Starting Facebook Reel Scheduling Workflow")

    // 1. Load pending posts
    state, err := w.repo.Load()
    if err != nil {
        return fmt.Errorf("failed to load scheduling state: %w", err)
    }

    // Count pending
    pendingCount := 0
    for _, p := range state.Posts {
        if !p.SchedulingInfo.IsSuccess {
            pendingCount++
        }
    }
    if pendingCount == 0 {
        logger.Info("FbReelsWorkflow:Run", "All posts are already marked as scheduled. Nothing to do.")
        return nil
    }
    logger.Info("FbReelsWorkflow:Run", fmt.Sprintf("Found %d pending reels to schedule", pendingCount))

    // 2. Launch Browser Session
    session, err := browser.NewSession(w.cfg.CookiesPath, headed)
    if err != nil {
        return fmt.Errorf("failed to init browser session: %w", err)
    }
    defer session.Close()

    page, err := session.Context.NewPage()
    if err != nil {
        return fmt.Errorf("failed to create browser page: %w", err)
    }

    fbPage := pages.NewFbSchedulingPage(page, w.cfg.FbBaseURL, w.cfg.FbScheduledContentURL)

    // Step 1: Open Page URL
    if err := fbPage.Open(state.FbPageSlug); err != nil {
        return fmt.Errorf("failed to open page slug: %w", err)
    }

    // Step 2: Handle Login & OTP pause if required
    time.Sleep(3 * time.Second)
    if err := fbPage.HandleLoginIfRequired(state.FbPageSlug, state.PageName); err != nil {
        return fmt.Errorf("login handling failed: %w", err)
    }

    // Step 3: Ensure on exact page and persist auth cookies
    if err := fbPage.WaitForExactPageURL(state.FbPageSlug); err != nil {
        return fmt.Errorf("failed verifying page URL: %w", err)
    }
    _ = session.SaveCookies()

    // Step 4: Navigate to Scheduled Content Library
    if err := fbPage.NavigateToScheduledContent(); err != nil {
        return fmt.Errorf("failed navigating to Content Library: %w", err)
    }

    page.WaitForTimeout(5000)

    // Step 5: Process each post
    for i := range state.Posts {
        post := &state.Posts[i]
        if post.SchedulingInfo.IsSuccess {
            logger.Info("FbReelsWorkflow:Run", fmt.Sprintf("Skipping already scheduled post: %s", post.VideoPath))
            continue
        }

        // Determine absolute video path; if post.VideoPath is already absolute, use it directly, otherwise join with WorkspaceRoot
    videoAbsPath := post.VideoPath
    if !filepath.IsAbs(videoAbsPath) {
        videoAbsPath = filepath.Join(w.cfg.WorkspaceRoot, videoAbsPath)
    }
    videoAbsPath = filepath.Clean(videoAbsPath)
        if _, err := os.Stat(videoAbsPath); os.IsNotExist(err) {
            logger.Error("FbReelsWorkflow:Run", fmt.Sprintf("Video file not found at %s. Marking tried=true", videoAbsPath))
            post.SchedulingInfo.Tried = true
            _ = w.repo.Save(state)
            continue
        }

        logger.Info("FbReelsWorkflow:Run", fmt.Sprintf("Scheduling video: %s", videoAbsPath))

        // Click Create
        if err := fbPage.ClickCreateButton(); err != nil {
            return fmt.Errorf("failed clicking Create button: %w", err)
        }

        // Upload Reel video
        if err := fbPage.UploadReelVideo(videoAbsPath); err != nil {
            return fmt.Errorf("failed uploading reel video: %w", err)
        }

        // Fill Caption and Submit
        if err := fbPage.FillDetailsAndSchedule(post.Caption); err != nil {
            return fmt.Errorf("failed completing reel scheduling: %w", err)
        }

        // Update State and Persist immediately
        post.SchedulingInfo.IsSuccess = true
        post.SchedulingInfo.Tried = true
        if err := w.repo.Save(state); err != nil {
            logger.Error("FbReelsWorkflow:Run", "Failed to persist updated state: "+err.Error())
        } else {
            logger.Info("FbReelsWorkflow:Run", fmt.Sprintf("Successfully persisted state for: %s", post.VideoPath))
        }

        page.WaitForTimeout(10000)
    }

    logger.Info("FbReelsWorkflow:Run", "All reels scheduled successfully")
    return nil
}
