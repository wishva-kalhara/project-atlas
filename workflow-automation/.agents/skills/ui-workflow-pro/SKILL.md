---
name: ui-workflow-pro
description: |
  A comprehensive Antigravity skill that bundles the entire `ui-workflows` Playwright automation project.
  It includes source code, configuration, documentation, logger enhancements, and test scripts.
  Loading this skill restores a ready‑to‑run UI‑automation suite for Google Flow and YouTube.
---

# Overview
The **ui‑workflow‑pro** skill packages the `ui-workflows` project located at:

```
<repo-root>/ui-workflows
```

It contains:
- TypeScript source (`src/`), configuration (`src/config/Config.ts`), utilities (`src/utils/`), and feature folders (`videoGenGoogleFlow`, `openYoutube`).
- Playwright test specifications.
- Documentation (`README.md`, `ARCHITECTURE.md`).
- Logger implementation (`src/utils/Logger.ts`) with slug‑based tracing.
- Cookie persistence (`cookies.json` placeholder).
- Development dependencies (`@playwright/test`, `@types/node`, `dotenv`, `typescript`).

# Installation
1. **Navigate to the project**
   ```bash
   cd d:\content-creation\workflow-automation\ui-workflows
   ```
2. **Install dependencies** (if not already installed)
   ```bash
   npm install
   ```
   > If you encounter the *"Cannot find name 'process'"* TypeScript error, run:
   ```bash
   npm i -D @types/node
   ```
3. **Create a `.env` file** (git‑ignored) with your Google Flow credentials:
   ```dotenv
   GOOGLE_FLOW_EMAIL=your.email@example.com
   GOOGLE_FLOW_PASSWORD=yourPassword
   ```

# Running the Automations
- Execute all tests:
  ```bash
  npm test
  ```
- Run a single feature (e.g., YouTube):
  ```bash
  npx playwright test -g "Open YouTube"
  ```
- The first run will generate `cookies.json`. Subsequent runs reuse the saved cookies, avoiding repeated logins.

# Logger Usage
All page‑object classes now import `Logger` and emit debug/info/warn messages with a **slug** (e.g., `LoginPage:fillEmail`).
Search the logs for a slug to pinpoint execution points.

# Skill Loading
To load this skill in an Antigravity session:
```text
define_subagent {
  Name: "ui-workflow-pro-loader",
  Description: "Loads the ui‑workflow‑pro skill so agents can reference its files.",
  SystemPrompt: "You have access to the ui‑workflow‑pro skill files under .agents/skills/ui-workflow-pro. Use them when asked about UI automation.",
  EnableWriteTools: false,
  EnableMCPTools: true,
  EnableSubagentTools: false
}
```
Then invoke it with:
```
invoke_subagent {
  Role: "UI Automation Skill Loader",
  Prompt: "Load the ui‑workflow‑pro skill and be ready to answer any questions about the UI automation suite.",
  TypeName: "ui-workflow-pro-loader"
}
```
The skill will be available for future sub‑agents or actions.

# Maintenance
- **Adding new automations**: create a new folder under `src/features/<newFeature>` following the existing feature layout.
- **Updating docs**: modify `README.md` and `ARCHITECTURE.md` as needed; the skill will always reference the latest versions.
- **Version bump**: update the `version` field in `package.json` and commit changes.

---
*All paths are absolute to the workspace root. The skill does **not** copy files; it references them directly, keeping the repository size minimal.*