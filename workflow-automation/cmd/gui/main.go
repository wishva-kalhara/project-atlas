package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/mxschmitt/playwright-go"
	"workflow-automation/internal/config"
	"workflow-automation/internal/data"
	"workflow-automation/internal/logger"
	"workflow-automation/internal/workflows"
)

type WorkflowOption struct {
	Name        string
	Description string
	Run         func(headed bool, index int) error
}

var menuOptions = []WorkflowOption{
	{
		Name:        "Schedule Facebook Reels",
		Description: "Schedule pending reels to Facebook Creator Dashboard",
		Run: func(headed bool, index int) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			repo := data.NewRepository(cfg.FbReelsDataPath)
			wf := workflows.NewFbReelsWorkflow(cfg, repo)
			return wf.Run(headed)
		},
	},
	{
		Name:        "Generate Video (Google Flow)",
		Description: "Run Google Flow video generation workflow",
		Run: func(headed bool, index int) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wf := workflows.NewGoogleFlowWorkflow(cfg)
			return wf.Run(headed, index)
		},
	},
	{
		Name:        "Open YouTube",
		Description: "Open YouTube in a browser session (verification helper)",
		Run: func(headed bool, index int) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			wf := workflows.NewYouTubeWorkflow(cfg)
			return wf.Run(headed)
		},
	},
	{
		Name:        "Show Status",
		Description: "Print current scheduling state for Facebook reels",
		Run: func(headed bool, index int) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			repo := data.NewRepository(cfg.FbReelsDataPath)
			state, err := repo.Load()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SLUG\tPAGE\tCAPTION\tVIDEO PATH\tTRIED\tSUCCESS")
			for _, p := range state.Posts {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%t\n",
					state.FbPageSlug,
					state.PageName,
					p.Caption,
					p.VideoPath,
					p.SchedulingInfo.Tried,
					p.SchedulingInfo.IsSuccess,
				)
			}
			w.Flush()
			return nil
		},
	},
	{
		Name:        "Setup Playwright",
		Description: "Download Playwright driver and Chromium browser",
		Run: func(headed bool, index int) error {
			return playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
		},
	},
}

func main() {
	headed := true
	if v := os.Getenv("WORKFLOW_EXEC_HEADLESS"); v == "1" || strings.EqualFold(v, "true") {
		headed = false
	}

	// Ask for index for video-gen
	index := 0

	clearScreen()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          WORKFLOW EXECUTOR - Interactive Menu               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println()

	for i, wf := range menuOptions {
		fmt.Printf("  %d. %s\n", i+1, wf.Name)
		fmt.Printf("     %s\n\n", wf.Description)
	}

	fmt.Printf("  %d. Exit\n\n", len(menuOptions)+1)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Print("\nSelect an option (1-", len(menuOptions)+1, "): ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("\nNo selection made. Exiting...")
		return
	}

	choice := 0
	_, err := fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(menuOptions)+1 {
		fmt.Println("\nInvalid selection. Please enter a number between 1 and", len(menuOptions)+1)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		return
	}

	if choice == len(menuOptions)+1 {
		fmt.Println("\nExiting...")
		return
	}

	// If video-gen selected, ask for index
	if choice == 2 { // Generate Video (Google Flow)
		fmt.Print("\nEnter metadata index (default 0): ")
		var indexInput string
		fmt.Scanln(&indexInput)
		indexInput = strings.TrimSpace(indexInput)
		if indexInput != "" {
			_, err := fmt.Sscanf(indexInput, "%d", &index)
			if err != nil {
				fmt.Println("Invalid index, using default 0")
				index = 0
			}
		}
	}

	selected := menuOptions[choice-1]
	fmt.Printf("\nRunning: %s\n", selected.Name)
	fmt.Println(strings.Repeat("-", 60))

	if err := selected.Run(headed, index); err != nil {
		logger.Error("gui", err.Error())
		fmt.Printf("\nError: %v\n", err)
		waitForEnter()
		return
	}

	logger.Info("gui", fmt.Sprintf("%s completed", selected.Name))
	waitForEnter()
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func waitForEnter() {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}
