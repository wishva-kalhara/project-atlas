# Project Context for Future Agents

## Overview
The **ui‑workflows** repository contains a Playwright UI‑automation suite written in TypeScript. It automates two primary user journeys:
1. **Google Flow video generation** – login via Google Flow, create a new project, and optionally handle phone verification.
2. **YouTube opening** – navigate to YouTube homepage.

The automation follows a **feature‑based architecture** with a **Page‑Object Model (POM)**, centralised configuration, cookie persistence, and structured logging.

## Key Components
- `src/config/Config.ts` – loads environment variables (`GOOGLE_FLOW_EMAIL`, `GOOGLE_FLOW_PASSWORD`) and defines URLs and cookie path.
- `src/utils/Logger.ts` – simple logger emitting messages with a *slug* (`<ClassName>:<method>`), aiding traceability.
- `src/utils/CookieManager.ts` – saves and loads cookies to `cookies.json` to avoid repeated logins.
- `src/features/<feature>/pages/` – POM classes (`BasePage`, `FlowHomePage`, `LoginPage`, `ProjectPage`, `YoutubeHomePage`).
- `src/features/<feature>/tests/` – Playwright test specs (`videoGenGoogleFlow.spec.ts`, `openYoutube.spec.ts`).
- `playwright.config.ts`, `tsconfig.json`, `package.json` – standard Playwright/TS setup.
- Documentation: `README.md` (quick‑start) and `ARCHITECTURE.md` (folder diagram, design decisions).

## Skill Packaging
The **ui‑workflow‑pro** skill (`.agents/skills/ui-workflow-pro/SKILL.md`) bundles this entire project, allowing future agents to load the automation suite with a single sub‑agent definition.

## Usage Workflow
1. `cd d:\content-creation\workflow-automation\ui-workflows`
2. `npm install` (or `npm i -D @types/node` if the *process* TS error occurs)
3. Create a `.env` with Google Flow credentials.
4. Run tests:
   - All: `npm test`
   - Specific: `npx playwright test -g "Open YouTube"`
5. Logs include slugs for precise location identification.

## Extensibility
- Add new automations by creating `src/features/<newFeature>` with its own `pages/` and `tests/`.
- Update `README.md` and `ARCHITECTURE.md` accordingly.
- Increment version in `package.json` when releasing changes.

---
*This file is designed for agents to quickly understand the structure, purpose, and usage of the UI automation project.*