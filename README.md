# things

A command-line interface for interacting with [Things.app](https://culturedcode.com/things/) tasks.

## Installation

```bash
go install github.com/mybuddymichael/things@latest
```

## Commands

- `show` - List to-dos from a specific list
- `add` - Create a new to-do in a list
- `delete` - Remove a to-do by name
- `move` - Move a to-do between lists
- `rename` - Rename a to-do
- `log` - View completed to-dos from the Logbook

## Usage

```bash
# View commands and help
things -h
things show -h

# Show to-dos in a list
things show --list "Today"

# Add a to-do with tags
things add --name "Review PR" --list "Work" --tags "urgent, code-review"

# View completed to-dos from today
things log --date today

# Filter completed to-dos by project
things log --date "this week" --project "Redesign"

# Output as JSONL for scripting
things show --list "Today" --jsonl
```
