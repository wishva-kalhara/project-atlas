package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	stateFile     = ".atlas_state"
	defaultGemURL = "https://gemini.google.com/"
)

// Step represents a single CLI execution stage in the E2E workflow
type Step struct {
	Name        string
	Bin         string
	Args        []string
	PromptInput bool // True if this step requires interactive user input before running
}

// Define the 7 sequential workflow steps
var fbSchedulingSteps = []Step{
	{
		Name: "Initialize Facebook Scheduling Workflow",
		Bin:  "./tools.exe",
		Args: []string{"new-workflow-fbScheduling"},
	},
	{
		Name:        "Run Gemini Gem Harness",
		Bin:         "./harness.exe",
		Args:        []string{"-url", defaultGemURL},
		PromptInput: true,
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

func atlasTheme() *huh.Theme {
	t := huh.ThemeCharm()
	blue := lipgloss.Color("#00A8FF")
	t.Focused.Base = t.Focused.Base.BorderForeground(blue)
	t.Focused.Title = t.Focused.Title.Foreground(blue)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(blue)
	return t
}

func printBanner() {
	fmt.Print("\033[H\033[2J")

	logo := `
   ___  ________   ___   ____
  / _ |/_  __/ /  / _ | / __/
 / __ | / / / /__/ __ |_\ \  
/_/ |_|/_/ /____/_/ |_/___/  
`

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00A8FF"))

	subtitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#04B575")).
		MarginTop(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00A8FF")).
		Padding(0, 3).
		MarginBottom(1)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(logo),
		subtitleStyle.Render("E2E WORKFLOW ORCHESTRATOR"),
	)

	fmt.Println(boxStyle.Render(content))
}

// getExeDir returns the directory containing atlas.exe so paths work from anywhere
func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func main() {
	for {
		printBanner()

		var selectedFlow string

		// 1. Main Dashboard Menu
		menuForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select an action:").
					Options(
						huh.NewOption("Generate Videos and Schedule to Facebook", "fb_schedule"),
						huh.NewOption("Clear Current Progress", "clear_progress"),
						huh.NewOption("Exit", "exit"),
					).
					Value(&selectedFlow),
			),
		).WithTheme(atlasTheme())

		err := menuForm.Run()
		if err != nil || selectedFlow == "exit" {
			fmt.Println("Exiting Atlas. Goodbye!")
			os.Exit(0)
		}

		switch selectedFlow {
		case "fb_schedule":
			runFbSchedulingFlow()
		case "clear_progress":
			clearCurrentProgress()
		}
	}
}

func clearCurrentProgress() {
	exeDir := getExeDir()

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Clear Current Progress?").
				Description("This will run './tools.exe clear-tmp-data' and reset your workflow progress to Step 1.").
				Affirmative("Yes, Clear Progress").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithTheme(atlasTheme())

	if err := confirmForm.Run(); err != nil || !confirm {
		return
	}

	fmt.Println("\n>>> Running: ./tools.exe clear-tmp-data")
	fmt.Println(strings.Repeat("-", 60))

	if err := executeCommand("./tools.exe", []string{"clear-tmp-data"}); err != nil {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("\n[Error] Failed to clear temporary data: %v\n", err)
		pauseForEnter()
		return
	}

	// Remove the step state tracker so the next run starts fresh at Step 1
	_ = os.Remove(filepath.Join(exeDir, stateFile))

	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\n✓ Current progress cleared. Next run will start from Step 1.")
	pauseForEnter()
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
		// Copy args slice so dynamic edits don't mutate the global template
		currentArgs := append([]string{}, step.Args...)

		// If this is the Gemini Harness step, ask for the URL before confirming execution
		if step.PromptInput {
			gemURL := defaultGemURL
			urlForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter Gemini Gem URL").
						Description("Press Enter to use the default URL or paste a new one:").
						Placeholder(defaultGemURL).
						Value(&gemURL),
				),
			).WithTheme(atlasTheme())

			if err := urlForm.Run(); err != nil {
				fmt.Printf("\nProgress paused at Step %d. Run atlas.exe again to resume.\n", i+1)
				os.Exit(0)
			}

			gemURL = strings.TrimSpace(gemURL)
			if gemURL == "" {
				gemURL = defaultGemURL
			}
			currentArgs = []string{"-url", gemURL}
		}

		cmdPreview := fmt.Sprintf("%s %s", step.Bin, strings.Join(currentArgs, " "))
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
		).WithTheme(atlasTheme())

		err := stepForm.Run()
		if err != nil || action == "stop" {
			fmt.Printf("\nProgress paused at Step %d. Run atlas.exe again to resume.\n", i+1)
			os.Exit(0)
		}

		// 3. Execute the command and stream stdout/stderr live
		fmt.Printf("\n>>> Running: %s\n", cmdPreview)
		fmt.Println(strings.Repeat("-", 60))

		if err := executeCommand(step.Bin, currentArgs); err != nil {
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
			).WithTheme(atlasTheme())
			_ = retryForm.Run()
			if retry {
				i--
				continue
			}
			os.Exit(1)
		}

		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("✓ Step %d/%d completed successfully!\n\n", i+1, len(fbSchedulingSteps))

		// 4. Update state after successful execution
		if i == len(fbSchedulingSteps)-1 {
			_ = os.Remove(filepath.Join(getExeDir(), stateFile))
			fmt.Println("🎉 All E2E workflow steps completed and temporary data cleared!")
			pauseForEnter()
		} else {
			saveStepState(i + 1)
		}
	}
}

func executeCommand(bin string, args []string) error {
	exeDir := getExeDir()
	cleanBin := filepath.Join(exeDir, filepath.Base(bin))

	cmd := exec.Command(cleanBin, args...)
	cmd.Dir = exeDir
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1", "FORCE_COLOR=1")

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	tag := filepath.Base(bin)
	outPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("#00A8FF")).Render(fmt.Sprintf("[%s] ", tag))
	errPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Render(fmt.Sprintf("[%s ERR] ", tag))

	var wg sync.WaitGroup
	wg.Add(2)

	streamLogs := func(r io.Reader, prefix string, out io.Writer) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			fmt.Fprintf(out, "%s%s\n", prefix, scanner.Text())
		}
	}

	go streamLogs(stdoutPipe, outPrefix, os.Stdout)
	go streamLogs(stderrPipe, errPrefix, os.Stderr)

	wg.Wait()
	return cmd.Wait()
}

func hasDataFolders() bool {
	pattern := filepath.Join(getExeDir(), ".data*")
	matches, err := filepath.Glob(pattern)
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

func detectCurrentStep() int {
	statePath := filepath.Join(getExeDir(), stateFile)
	if !hasDataFolders() {
		_ = os.Remove(statePath)
		return 0
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return 1
	}

	stepIdx, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || stepIdx < 0 || stepIdx >= len(fbSchedulingSteps) {
		return 0
	}

	return stepIdx
}

func saveStepState(nextStepIndex int) {
	statePath := filepath.Join(getExeDir(), stateFile)
	_ = os.WriteFile(statePath, []byte(strconv.Itoa(nextStepIndex)), 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func pauseForEnter() {
	fmt.Print("\nPress Enter to return to the main dashboard...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
