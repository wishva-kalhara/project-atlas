# Tools

A small Go CLI that bridges the Gemini harness output into the Google Flow video-generation input format.

## What it does

The Gemini harness (`../harness`) writes one `.md` file per prompt tag into
`.data-harness-gemini-out/`. Each `.md` contains a full video-generation prompt
starting at the `FORMAT AND INTENT:` marker. This tool converts that output into
the JSON contract expected by the Google Flow workflow:

1. Reads `.data-harness-gemini-in/human-prompts.json` for the list of tags and
   their associated images.
2. For each tag, reads `.data-harness-gemini-out/<tag>.md` and extracts
   everything from `FORMAT AND INTENT:` to the end of the file as the
   `videoPrompt`.
3. Writes `.data-workflow-googleflow-in/metadata.json` in this shape:

```json
[
  {
    "tag": "11-20-1",
    "images": ["./image.png"],
    "videoPrompt": "FORMAT AND INTENT: ..."
  }
]
```

## Folder layout

```
tools/
├── main.go                     # Tool source
├── go.mod / go.sum             # Go module
├── tools.exe                   # Built binary (git-ignored)
├── README.md
├── .data-harness-gemini-in/    # human-prompts.json + images (git-ignored)
├── .data-harness-gemini-out/   # <tag>.md files from the Gemini harness (git-ignored)
└── .data-workflow-googleflow-in/   # metadata.json output (git-ignored)
```

## Prerequisites

- **Go 1.24+** (`go version`)
- Nothing else — the tool only uses the Go standard library.

## Build

```bash
go build -o tools.exe .
```

## Run

The tool is command-driven. Running it with no matching command prints
`no action binded` and exits:

```bash
./tools.exe                 # -> "no action binded"
./tools.exe gen-workflow-googleflow-in
```

`gen-workflow-googleflow-in` reads the three folders relative to its own working
directory, so run it from `tools/` (or pass the paths you want via flags if you
ever add them).

## Re-running

The tool is idempotent: it regenerates `.data-workflow-googleflow-in/metadata.json`
from the current `.md` files each time. Re-run it whenever the Gemini harness
produces new or updated `.md` files.