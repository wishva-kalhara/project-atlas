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

## Commands

The tool is command-driven. Running it with no matching command prints
`no action binded` and exits:

```bash
./tools.exe                 # -> "no action binded"
./tools.exe gen-workflow-googleflow-in
./tools.exe gen-workflow-fbSchedule-in --fbPageSlug <slug> --pageName "<name>"
./tools.exe clear-tmp-data
```

### `gen-workflow-googleflow-in`

Reads `.data-harness-gemini-in/human-prompts.json` for the list of tags and
their associated images, extracts the `FORMAT AND INTENT:` prompt from each
`.data-harness-gemini-out/<tag>.md`, and writes
`.data-workflow-googleflow-in/metadata.json`:

```json
[
  {
    "tag": "11-23-1",
    "images": ["./image.png"],
    "videoPrompt": "FORMAT AND INTENT: ..."
  }
]
```

### `gen-workflow-fbSchedule-in`

Reads `.data-harness-gemini-in/human-prompts.json` for the list of tags, copies
the `CAPTION` section from each `.data-harness-gemini-out/<tag>.md` into the
`caption` field of a post, and points each post at its rendered video at
`{{current_path}}/.data-workflow-googleflow-out/<tag>.mp4`. Writes
`.data-workflow-fbSchedule-in/metadata.json` in the shape consumed by
`workflow-automation`:

```json
{
  "fbPageSlug": "marine.masterpieces",
  "pageName": "Marine masterpieces",
  "posts": [
    {
      "caption": "Watch This Peacock Mantis Shrimp Stand Its Ground 🦐 ...",
      "videoPath": "D:/Projects/project-atlas/tools/.data-workflow-googleflow-out/11-23-1.mp4",
      "schdulingInfo": {
        "isSuccess": false,
        "tried": false
      }
    }
  ]
}
```

Both `--fbPageSlug` and `--pageName` are required.

### `clear-tmp-data`

Deletes all `.data*` directories in the working directory, leaving only source files and the binary. Use this to reset the workspace to a clean state.

```bash
./tools.exe clear-tmp-data
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
├── .data-workflow-googleflow-in/   # metadata.json output (git-ignored)
└── .data-workflow-fbSchedule-in/    # metadata.json output (git-ignored)
```

## Prerequisites

- **Go 1.24+** (`go version`)
- Nothing else — the tool only uses the Go standard library.

## Build

```bash
go build -o tools.exe .
```

## Re-running

The tool is idempotent: it regenerates its `metadata.json` from the current
`.md` files each time. Re-run it whenever the Gemini harness produces new or
updated `.md` files.