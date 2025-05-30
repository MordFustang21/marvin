# Custom Commands for Marvin

This directory contains example JSON configuration files for Marvin's custom commands feature. Custom commands allow you to define your own shortcuts for common tasks, programs, and websites directly in Marvin's search interface.

## Getting Started

1. Create a directory to store your command definitions:
   ```
   mkdir -p ~/.config/marvin/commands
   ```

2. Copy the example JSON files to your commands directory:
   ```
   cp git-commands.json ~/.config/marvin/commands/
   cp system-tools.json ~/.config/marvin/commands/
   ```

3. Create a directory for your custom icons (optional):
   ```
   mkdir -p ~/.config/marvin/commands/icons
   ```

4. Add your own icons to the icons directory (optional)

5. Restart Marvin to load your custom commands

## Creating Your Own Commands

You can create your own command definition files by following these steps:

1. Create a new JSON file in `~/.config/marvin/commands/` (e.g., `my-commands.json`)
2. Follow the structure in the example files
3. Restart Marvin after adding new command files

## JSON Structure

Each JSON file defines a group of related commands with the following structure:

```json
{
  "name": "Group Name",
  "description": "Group description",
  "icon": "path/to/icon.png",
  "commands": [
    {
      "name": "Command Name",
      "trigger": "search text",
      "description": "What this command does",
      "action": {
        "type": "shell|url|application",
        "command": "shell command",
        "url": "https://example.com",
        "path": "/Applications/Example.app"
      }
    }
  ]
}
```

### Field Descriptions

#### Command Provider
- `name`: Name of the command group
- `description`: Description of the command group
- `icon`: Path to an icon file (optional). Can be relative to the commands directory or absolute.
- `commands`: Array of individual commands

#### Command
- `name`: Name of the command (shown in search results)
- `trigger`: Text that triggers this command in the search
- `description`: Description shown in search results
- `action`: The action to perform
- `icon`: Individual command icon (optional, overrides group icon)

#### Action Types
- `shell`: Executes a shell command
  - `command`: The shell command to execute
- `url`: Opens a URL in the default browser
  - `url`: The URL to open
- `application`: Opens an application
  - `path`: Path to the application

## Examples

### Git Commands
The `git-commands.json` file contains shortcuts for common Git operations. Once installed, you can type "git" in Marvin to see all available Git commands, or "git pull" to run that specific command.

### System Tools
The `system-tools.json` file contains various system utilities. For example, typing "disk" will show disk usage information, and "preferences" will open System Preferences.

## Tips

1. Use short, memorable trigger words
2. Group related commands in the same file
3. Use descriptive names and descriptions
4. Consider adding custom icons for better visual recognition
5. For shell commands that need to run in a specific directory, include `cd /path/to/dir &&` at the start of the command