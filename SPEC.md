# Grimmoir Technical Specification

Version: 0.1.0 (MVP)

## 1. Project Overview

Grimmoir is a Go CLI/TUI for managing local Markdown prompts and agent skills.

- It treats a local folder of `.md` files as the source of truth.
- It supports both interactive usage (TUI) and scriptable usage (CLI).
- It integrates with Git for syncing prompt files.

## 2. Data Architecture

### Storage

- Default path: `~/.grimmoir`
- Override path: `GRIMMOIR_PATH` or `--path`
- Format: Markdown files with YAML frontmatter

### Skill File Format

Example: `~/.grimmoir/security_auditor.md`

```markdown
---
name: "Security Auditor"
description: "Audits Go code for common vulnerabilities."
tags: [golang, security, review]
version: "1.0.0"
---

# Security Auditor Skill
You are an expert security researcher. Review the following code for:
1. SQL Injection
2. Buffer Overflows
```

## 3. Core Features

### A. CLI (Automation Mode)

- `grim list`: list all indexed skills in a table
- `grim get <name>`: print raw prompt body to stdout
- `grim add --name "..." --body "..."`: add skill from text
- `grim add --name "..." --clip`: add skill from clipboard
- `grim delete "..."`: delete skill by name
- `grim sync`: sync prompt markdown files via Git
- `grim tui`: launch interactive TUI

### B. TUI (Interactive Mode)

- Real-time search over name, tags, description (`/`)
- Prompt stack selection (`c`) and compose+copy (`enter`)
- Delete selected prompt (`d`)
- Create new prompt from clipboard (`n`)
- Live preview pane for selected prompt
- Responsive layout for narrow and wide terminal sizes
- Always-visible command guide/header

### C. Compose Behavior

- Composition order follows stack order
- Concatenation uses H1 boundaries (`# <name>`) between prompts
- Result is copied to clipboard

## 4. Git Sync Behavior

`grim sync` runs in the configured prompt store path (not shell cwd).

Flow:

1. Ensure Git repository exists in store path (`git rev-parse`, then `git init` if missing)
2. Stage prompt files only: `git add -- *.md`
3. Commit with timestamp: `Grimmoir sync: <RFC3339 time>`
4. `git pull --rebase`
5. `git push`

Notes:

- Store path must be configured to a directory that should be version-controlled.
- Remote/upstream must exist for push to succeed.

## 5. Stack & Dependencies

- Language: Go
- CLI: Cobra
- TUI: Bubble Tea + Lip Gloss
- Clipboard: atotto/clipboard
- Frontmatter parsing: `gopkg.in/yaml.v3`

## 6. Quality Requirements

- Red/green TDD for behavior changes
- Predictable error messages for user-facing failures
- Keep implementation simple and local-first
- Maintain docs/spec alignment with current behavior
