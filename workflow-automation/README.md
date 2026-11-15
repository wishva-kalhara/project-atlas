# Workflow Automation

A Go CLI for automating content workflows — Facebook Reel scheduling, Google Flow video generation, and YouTube session management — powered by [Playwright](https://playwright.dev/) browser automation.

## Prerequisites

- **Go** 1.24+
- **Google Chrome** installed (the CLI uses your system Chrome; falls back to bundled Chromium)

## Quick Start

```bash
# 1. Clone the repo
git clone <repo-url> && cd workflow-automation

# 2. Build the binaries
go build -o workflow-exec.exe .
go build -o workflow-exec-gui.exe ./cmd/gui

# 3. Install Playwright driver & browsers
./workflow-exec setup

# 4. Run a workflow
./workflow-exec schedule-fb --headed
```

### Double-Click to Run (Interactive GUI)

`workflow-exec-gui.exe` is a GUI-friendly entry point. When you double-click it, it shows a numbered menu, you pick a workflow, and it **directly triggers the Playwright automations** — no separate process, no shell-out to `workflow-exec.exe`. It shares the same `internal/workflows` package, so the menu options are always in sync with the CLI commands.

```bash
# Just build and double-click:
go build -o workflow-exec-gui.exe ./cmd/gui
```

Set `WORKFLOW_EXEC_HEADLESS=1` to run the browser headless instead of headed.

## Commands

| Command | Description |
| --- | --- |
| `setup` | Downloads the Playwright driver and Chromium browser |
| `schedule-fb` | Schedules pending reels to Facebook Creator Dashboard |
| `video-gen` | Runs the Google Flow video generation workflow |
| `youtube` | Opens YouTube in a managed browser session |
| `status` | Prints the current scheduling state for Facebook reels |
| `gui` | Interactive menu that directly triggers any of the above |

### Global Flags

| Flag | Default | Description |
| --- | --- | --- |
| `--headed` | `true` | Run browser with a visible UI |
| `--debug` | `false` | Enable verbose debug logs |

## Configuration

The CLI loads a `.env` file from the working directory (if present).

| Variable | Purpose |
| --- | --- |
| `GOOGLE_FLOW_EMAIL` | Google account email for Flow sign-in |
| `GOOGLE_FLOW_PASSWORD` | Google account password for Flow sign-in |

Scheduling data lives in `.data/scheduling-engine/fb-reels.json`. Browser cookies are persisted to `cookies.json` in the workspace root.

## Project Structure

```
.
├── main.go                          # Entry point
├── cmd/
│   ├── root.go                      # Cobra root command & global flags
│   ├── commands.go                  # Subcommand definitions (schedule-fb, video-gen, youtube, setup)
│   ├── status.go                    # Status subcommand
│   └── gui/                         # Interactive GUI entry point
│       └── main.go                  # Builds workflow-exec-gui.exe
│   └── wrapper/                     # Interactive wrapper that selects & runs workflows
│       └── main.go                  # Builds workflow-exec-wrapper.exe
├── internal/
│   ├── browser/
│   │   └── session.go               # Playwright session lifecycle (launch, cookies, teardown)
│   ├── config/
│   │   └── config.go                # .env loading & Config struct
│   ├── data/
│   │   ├── models.go                # JSON data models (Post, SchedulingInfo)
│   │   └── repository.go            # Read/write .data JSON files
│   ├── logger/
│   │   └── logger.go                # Structured logger
│   ├── pages/
│   │   ├── base_page.go             # Base page object
│   │   └── fb_scheduling_page.go    # Facebook scheduling page object
│   └── workflows/
│       ├── fb_reels.go              # Facebook Reels scheduling workflow
│       ├── google_flow.go           # Google Flow video generation workflow
│       └── youtube.go               # YouTube session workflow
├── .data/                           # Runtime data (git-ignored)
├── go.mod
└── go.sum
```

## Building from Source

```bash
go build -o workflow-exec.exe .
```

> The binary is **git-ignored** — CI builds it automatically (see below).

## CI / CD

A GitHub Actions workflow builds the binary on every push or PR to `main`. The compiled artifact is uploaded and available for download from the Actions run.

The workflow file lives at `.github/workflows/build.yml`.
