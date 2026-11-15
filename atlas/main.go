package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const stateFile = ".atlas_state"

// Step represents a single CLI execution stage in the E2E workflow
type Step struct {
	Name string
	Bin  string
	Args []string
}

// Define the 7 sequential workflow steps
var fbSchedulingSteps = []Step{
	{
		Name: "Initialize Facebook Scheduling Workflow",
		Bin:  "./tools.exe",
		Args: []string{"new-workflow-fbScheduling"},
	},
	{
		Name: "Run Gemini Gem Harness",
		Bin:  "./harness.exe",
		Args: []string{"-url", "https://gemini.google.com/gem/9bfe84b773a7/a4f05b58db4c8827"},
	},
	{
		Name: "Generate Google Flow Input",
		Bin:  "./tools.exe",
		Args: []string{"gen-workflow-googleflow-in"},
	},
	{
		Name: "Execute Video Generation (Headed)",
		Bin:  "./workflow-exec.exe",
		Args: []string{"video-gen", "--headed", "--index=0"},
	},
	{
		Name: "Generate Facebook Schedule Input",
		Bin:  "./tools.exe",
		Args: []string{"gen-workflow-fbSchedule-in", "--fbPageSlug", "marine.masterpieces", "--pageName", "Marine Masterpieces"},
	},
	{
		Name: "Execute Facebook Scheduling (Headed)",
		Bin:  "./workflow-exec.exe",
		Args: []string{"schedule-fb", "--headed"},
	},
	{
		Name: "Clear Temporary Data",
		Bin:  "./tools.exe",
		Args: []string{"clear-tmp-data"},
	},
}

func printBanner() {
	// Clear terminal screen for a clean CLI start
	fmt.Print("\033[H\033[2J")

	logo := `
   ___  ________   ___   ____
  / _ |/_  __/ /  / _ | / __/
 / __ | / / / /__/ __ |_\ \  
/_/ |_|/_/ /____/_/ |_/___/  
`

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	subtitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#04B575")).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 3).
		MarginBottom(1)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(logo),
		subtitleStyle.Render("E2E WORKFLOW ORCHESTRATOR"),
	)

	fmt.Println(boxStyle.Render(content))
}

func main() {
	for {
		printBanner()

		var selectedFlow string

		// 1. Main Menu Prompt
		menuForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an end-to-end flow to execute:").
					Options(
						huh.NewOption("Generate Videos and Schedule to Facebook", "fb_schedule"),
						huh.NewOption("Exit", "exit"),
					).
					Value(&selectedFlow),
			),
		)

		err := menuForm.Run()
		if err != nil || selectedFlow == "exit" {
			fmt.Println("Exiting Atlas. Goodbye!")
			os.Exit(0)
		}

		if selectedFlow == "fb_schedule" {
			runFbSchedulingFlow()
		}
	}
}

func runFbSchedulingFlow() {
	startIndex := detectCurrentStep()

	if startIndex > 0 {
		fmt.Printf("\n[Resume Detected] Existing .data folder found. Resuming from Step %d of %d.\n\n", startIndex+1, len(fbSchedulingSteps))
	} else {
		fmt.Printf("\n[Fresh Start] No active .data folders found. Starting from Step 1.\n\n")
	}

	for i := startIndex; i < len(fbSchedulingSteps); i++ {
		step := fbSchedulingSteps[i]
		cmdPreview := fmt.Sprintf("%s %s", step.Bin, strings.Join(step.Args, " "))

		var action string

		// 2. Stepper Confirmation Prompt before every step
		stepForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Step %d/%d: %s", i+1, len(fbSchedulingSteps), step.Name)).
					Description(fmt.Sprintf("Command to run:\n  %s", cmdPreview)).
					Options(
						huh.NewOption("Proceed (Execute this step)", "proceed"),
						huh.NewOption("Stop progress for now (Save & Exit)", "stop"),
					).
					Value(&action),
			),
		)

		err := stepForm.Run()
		if err != nil || action == "stop" {
			fmt.Printf("\nProgress paused at Step %d. Run atlas.exe again to resume.\n", i+1)
			os.Exit(0)
		}

		// 3. Execute the command and stream stdout/stderr live
		fmt.Printf("\n>>> Running: %s\n", cmdPreview)
		fmt.Println(strings.Repeat("-", 60))

		if err := executeCommand(step.Bin, step.Args); err != nil {
			fmt.Println(strings.Repeat("-", 60))
			fmt.Printf("\n[Error] Step %d failed: %v\n", i+1, err)

			var retry bool
			retryForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Command failed. Would you like to retry this step?").
						Affirmative("Retry Step").
						Negative("Exit Atlas").
						Value(&retry),
				),
			)
			_ = retryForm.Run()
			if retry {
				i-- // Decrement index to repeat current step
				continue
			}
			os.Exit(1)
		}

		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("✓ Step %d/%d completed successfully!\n\n", i+1, len(fbSchedulingSteps))

		// 4. Update state after successful execution
		if i == len(fbSchedulingSteps)-1 {
			// Last step (clear-tmp-data) completed; clean up state tracker
			_ = os.Remove(stateFile)
			fmt.Println("🎉 All E2E workflow steps completed and temporary data cleared!")
		} else {
			saveStepState(i + 1)
		}
	}
}

// executeCommand runs the target binary and pipes output directly to the user's terminal
func executeCommand(bin string, args []string) error {
	cleanBin := filepath.FromSlash(bin)

	cmd := exec.Command(cleanBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// hasDataFolders checks if any directory matching ".data*" exists in the working directory
func hasDataFolders() bool {
	matches, err := filepath.Glob(".data*")
	if err != nil || len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// detectCurrentStep determines which step index to start at based on .data* folders and .atlas_state
func detectCurrentStep() int {
	if !hasDataFolders() {
		_ = os.Remove(stateFile)
		return 0
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 1
	}

	stepIdx, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || stepIdx < 0 || stepIdx >= len(fbSchedulingSteps) {
		return 0
	}

	return stepIdx
}

// saveStepState writes the next step index to .atlas_state
func saveStepState(nextStepIndex int) {
	_ = os.WriteFile(stateFile, []byte(strconv.Itoa(nextStepIndex)), 0644)
}
