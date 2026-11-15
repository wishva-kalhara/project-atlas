package workflows

import (
    "workflow-automation/internal/browser"
    "workflow-automation/internal/config"
    "workflow-automation/internal/logger"
)

type YouTubeWorkflow struct {
    cfg *config.Config
}

func NewYouTubeWorkflow(cfg *config.Config) *YouTubeWorkflow {
    return &YouTubeWorkflow{cfg: cfg}
}

func (y *YouTubeWorkflow) Run(headed bool) error {
    logger.Info("YouTubeWorkflow:Run", "Navigating to YouTube")
    session, err := browser.NewSession(y.cfg.CookiesPath, headed)
    if err != nil {
        return err
    }
    defer session.Close()

    page, err := session.Context.NewPage()
    if err != nil {
        return err
    }

    if _, err := page.Goto(y.cfg.YouTubeURL); err != nil {
        return err
    }
    page.WaitForTimeout(10000)
    logger.Info("YouTubeWorkflow:Run", "YouTube loaded successfully")
    return nil
}
