# gtasks2md

A CLI tool written in Go to synchronize Google Tasks with local Markdown files. `gtasks2md` allows you to export your Google Tasks into structured Markdown and, vice versa, import and sync local Markdown checklists back up to Google Tasks.

## Features

- **Export:** Save all your Google Task lists to individual `.md` files in a directory, or export a single list directly to a specific `.md` file.
- **Import:** Push a directory of `.md` checklists to Google Tasks (matching lists by the H1 title in the Markdown), or upload a single `.md` file to a specified list.
- **Hierarchy support:** Natively supports subtasks and multiline notes.
- **Two-way Sync:** Automatically compares task states and updates the titles, notes, and completions accordingly when importing.

## Installation

You can download pre-built binaries for Linux, macOS, and Windows directly from the GitHub Releases page.

## Prerequisites

1. **Google Cloud Project** with the **Google Tasks API** enabled.
2. An OAuth 2.0 Client ID (`credentials.json`) configured as a "Desktop app".

## Setup

1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project or select an existing one.
3. Navigate to **APIs & Services > Library** and enable the **Google Tasks API**.
4. Go to **APIs & Services > Credentials**.
5. Click **Create Credentials > OAuth client ID**.
6. Select **Desktop app** as the application type.
7. Download the resulting JSON file and rename it to `credentials.json`.
8. Place `credentials.json` in the root of the project directory or configure the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to point to it.

## Building the Tool

Alternatively, you can clone the repository and build the binary from source using the Go toolchain. Use **Go 1.20.x** or newer.

```bash
# Clone the repository
git clone <your-repo-url>
cd google-tasks-import

# Build the executable
go build -o gtasks2md main.go
```

## Usage

### Authentication

On your very first run, `gtasks2md` will print a URL to your console. Open this URL in your browser, authenticate with your Google Account, and paste the authorization code back into the terminal. This will generate a local `token.json` file so you won't need to authenticate again until the token expires.

### Global Flags

- `-c, --credentials string`: Path to the OAuth 2.0 `credentials.json` file (default is `credentials.json` in the current directory or `GOOGLE_APPLICATION_CREDENTIALS`).

### Exporting Tasks

Export task lists from Google Tasks to local Markdown files.

```bash
# Export all your task lists into the current directory
./gtasks2md export

# Export all your task lists into a specific folder
./gtasks2md export ./my-tasks

# Export a single specific list to a single Markdown file
./gtasks2md export ./my-tasks/groceries.md --list-name "Groceries"
```

### Importing Tasks

Import task lists from local Markdown files up to Google Tasks.

```bash
# Import all markdown files in a directory
# The list name on Google Tasks will be determined by the `# Heading 1` of each file
./gtasks2md import ./my-tasks

# Import a single markdown file
./gtasks2md import ./my-tasks/groceries.md

# Import a single markdown file into a specifically named list (overrides the heading)
./gtasks2md import ./my-tasks/groceries.md --list-name "Weekend Shopping"
```

## Markdown Structure

The sync process relies on a specific structural format in your Markdown files. A valid Google Tasks list export looks like this:

```markdown
# My Google Tasks

- [ ] Buy groceries
    Milk, Eggs, Bread
    - [x] Pay at checkout
        Use new credit card
- [x] Clean the house
    Focus on living room
```

**Rules:**
- **Title:** The top `# H1` defines the Google Task List title.
- **Tasks:** Top-level tasks are defined using the `- [ ] ` or `- [x] ` checklist syntax.
- **Subtasks:** Must be indented with 4 spaces or a single tab under their parent task.
- **Notes:** Any text placed directly underneath a task/subtask and indented accordingly will be treated as the task's note.
