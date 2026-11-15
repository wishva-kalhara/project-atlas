package cmd

import (
	"fmt"

	"github.com/mxschmitt/playwright-go"
	"github.com/spf13/cobra"
	"workflow-automation/internal/config"
	"workflow-automation/internal/data"
	"workflow-automation/internal/logger"
	"workflow-automation/internal/workflows"
)

var scheduleFbCmd = &cobra.Command{
	Use:   "schedule-fb",
	Short: "Schedule pending reels to Facebook Creator Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		repo := data.NewRepository(cfg.FbReelsDataPath)
		wf := workflows.NewFbReelsWorkflow(cfg, repo)
		if err := wf.Run(headedFlag); err != nil {
			logger.Error("schedule-fb", err.Error())
			return err
		}
		logger.Info("schedule-fb", "Workflow completed")
		return nil
	},
}

var videoGenCmd = &cobra.Command{
	Use:   "video-gen",
	Short: "Run Google Flow video generation workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		wf := workflows.NewGoogleFlowWorkflow(cfg)
		if err := wf.Run(headedFlag, indexFlag); err != nil {
			logger.Error("video-gen", err.Error())
			return err
		}
		logger.Info("video-gen", "Workflow completed")
		return nil
	},
}

var youtubeCmd = &cobra.Command{
	Use:   "youtube",
	Short: "Open YouTube in a browser session (verification helper)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		wf := workflows.NewYouTubeWorkflow(cfg)
		if err := wf.Run(headedFlag); err != nil {
			logger.Error("youtube", err.Error())
			return err
		}
		logger.Info("youtube", "Workflow completed")
		return nil
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Bootstrap Playwright drivers (downloads Chromium if needed)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Playwright Go provides Install function
		fmt.Println("Installing Playwright browsers and driver…")
		if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
			return fmt.Errorf("playwright install failed: %w", err)
		}
		fmt.Println("Playwright setup complete.")
		return nil
	},
}
