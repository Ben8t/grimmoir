# Grimmoir

Grimmoir is a Go CLI/TUI for managing local markdown prompts and agent skills.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/Ben8t/grimmoir/main/scripts/install.sh | bash
```

Optional environment variables:

- `VERSION=v0.1.0` to install a specific release tag
- `INSTALL_DIR=$HOME/.local/bin` to override the install destination

## Storage

- Default path: `~/.grimmoir`
- Override with `GRIMMOIR_PATH` or `--path`
- File format: Markdown with YAML frontmatter

## Commands

- `grim list`: list indexed skills
- `grim get <name>`: print raw markdown body
- `grim add --name "Expert" --body "..."`: add a skill from inline text
- `grim add --name "Expert" --clip`: add a skill from clipboard contents
- `grim delete "Expert"`: delete a skill by name
- `grim sync`: in the configured store path repo, run `git add -- *.md`, commit with timestamp, `git pull --rebase`, `git push`
- `grim tui`: launch interactive mode

## TUI Keys

- `/`: search (name, tags, description)
- `j`/`k` or arrows: move selection
- `c`: add/remove selected skill from composition stack
- `d`: delete selected skill
- `enter`: compose stack and copy to clipboard
- `n`: create a new skill from clipboard (enter a name, press enter)
- `q`: quit

## Build

```bash
go build ./cmd/grim
```

## Test

```bash
go test ./...
```
