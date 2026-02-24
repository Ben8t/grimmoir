📜 Grimmoir: Technical SpecificationVersion: 0.1.0 (MVP)Stack: Go, Bubble Tea (TUI), Cobra (CLI)Storage: Local Markdown + Git1. Project OverviewGrimmoir is a CLI/TUI hybrid designed to manage, compose, and deploy LLM prompts and "Agent Skills." It treats a local directory of Markdown files as a version-controlled index, providing instant access via fuzzy search and clipboard integration.2. Data ArchitectureStorage StrategyRoot Directory: ~/.grimmoir/ (or configurable via GRIMMOIR_PATH).Format: Markdown files (.md) with YAML Frontmatter.Compatibility: Strictly follows the AgentSkills.io spec.Example File Structure (~/.grimmoir/code_review.md)Markdown---
name: "Security Auditor"
description: "Audits Go code for common vulnerabilities."
tags: [golang, security, review]
version: "1.0.0"
---

# Security Auditor Skill
You are an expert security researcher. Review the following code for:
1. SQL Injection
2. Buffer Overflows
...
3. Core Features & WorkflowA. The TUI (Interactive Mode)Built with Bubble Tea, providing a "Spotlight-like" experience for prompts.Search (/): Real-time fuzzy filtering across name, tags, and description.Preview: A side pane showing the content of the selected prompt.Composition Mode (c): * Select multiple prompts in sequence.Visual "stack" indicator showing selected items.Action: Press Enter to concatenate all selected prompts (separated by H1 headers) and copy to the system clipboard.New Prompt (n): Opens a buffer (or $EDITOR) to paste from clipboard and save.B. The CLI (Automation Mode)Designed for piping and shell scripts.grim get <name>: Outputs the raw prompt text to stdout.grim list: Lists all indexed skills/prompts in a table.grim add --name "Expert" --clip: Creates a new entry directly from the clipboard.C. Sync & Git Integrationgrim sync:git add . and git commit -m "Grimmoir sync: [Timestamp]"git pull --rebase (Manual merge required if conflicts occur).git push4. Technical Requirements (Go Stack)ComponentLibrary RecommendationTUI FrameworkCharm Bracelet Bubble TeaFuzzy FindingGo-fuzzyfinder or custom logic with bubbles/listCLI ArgumentsCobraClipboardatotto/clipboardFrontmatterHugo's go-yaml or yuin/goldmark5. Potential ConstraintsmacOS Permissions: Ensure the binary has permissions to access the clipboard and the specific directory.Git Auth: Relies on the user's SSH/HTTP helper being configured in the shell.
