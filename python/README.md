# Google Tasks to Markdown Sync

A command-line tool that syncs your Google Tasks with local Markdown files. It allows you to export your tasks into a readable Markdown format, edit them locally, and import them back into Google Tasks.

## Installation

1. Clone this repository or download the source code.
2. (Optional but recommended) Create and activate a Python virtual environment.
3. Install the dependencies using `pip`:

```bash
pip install -r requirements.txt
```

## Authentication

This tool uses the Google Tasks API, which requires OAuth 2.0 authentication.

### Obtaining Credentials
1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project or select an existing one.
3. Enable the **Google Tasks API** for your project.
4. Go to **APIs & Services > Credentials**.
5. Click **Create Credentials > OAuth client ID**.
6. Select **Desktop app** as the application type and create it.
7. Download the JSON file and rename it to `credentials.json`.

### Passing Credentials to the Tool
You can provide the credentials to the tool in three ways:
1. **Command Line Flag**: Use the `-c` or `--credentials` flag when running a command.
   ```bash
   python main.py export ./tasks --credentials path/to/credentials.json
   ```
2. **Environment Variable**: Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of your `credentials.json` file.
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="path/to/credentials.json"
   ```
3. **Default Location**: Place the `credentials.json` file in the current working directory where you run the tool.

*Note: The first time you run a command, a browser window will open to authenticate your Google account. A `token.json` file will be created to store your session locally for future use.*

## Usage Examples

The tool supports exporting and importing single task lists or multiple lists at once.

### Exporting Tasks

*   **Export a single list to a file:**
    ```bash
    python main.py export my_tasks.md --list-name "My Tasks"
    ```
*   **Export all lists to a directory:**
    ```bash
    python main.py export ./my_tasks_directory
    ```

### Importing Tasks

*   **Import a single file to a list:**
    ```bash
    python main.py import my_tasks.md --list-name "Target List Name"
    ```
    *(If `--list-name` is omitted, the title inside the markdown file or the filename is used.)*

*   **Import all files from a directory:**
    ```bash
    python main.py import ./my_tasks_directory
    ```

## Markdown Format Specification

The tool uses a strict Markdown format to map accurately to the Google Tasks API model. Google Tasks supports top-level tasks, notes (descriptions), completion status, and exactly one level of nesting (subtasks).

### Format Rules

1.  **List Title**: The first `H1` (`#`) in the document determines the Task List title.
2.  **Top-Level Tasks**: Unordered list items at the root indentation level with a checkbox (`- [ ]` or `- [x]`).
3.  **Completion Status**: Represented using standard task list syntax:
    *   `[ ]` for pending tasks
    *   `[x]` or `[X]` for completed tasks
4.  **Nested Tasks (Subtasks)**: Unordered list items indented by exactly 4 spaces (or 1 tab) under a parent task. *Note: Google Tasks API only supports 1 level of depth.*
5.  **Notes (Descriptions)**: Any text lines indented by exactly 4 spaces (or 1 tab) that do *not* start with a checkbox. These lines belong to the most recently declared task or subtask above them.

### Example `tasks.md`

```markdown
# My Tasks

- [ ] Buy groceries
    Need to get milk, eggs, and bread.
    Also check if they have any fresh strawberries.
    - [ ] Milk
    - [x] Bread
- [x] Call the plumber
    Phone number: 555-0198. Issue is the leaky kitchen sink.
- [ ] Schedule dentist appointment
```
