# HFL Documentation

Complete guide to using HFL for journaling and sync.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Date Parsing](#date-parsing)
- [Configuration](#configuration)
- [Notion Integration](#notion-integration)
- [Commands Reference](#commands-reference)
- [File Format](#file-format)
- [Troubleshooting](#troubleshooting)

## Installation

### Binary Release (Recommended)
1. Download the latest release from [GitHub releases](https://github.com/ahmaruff/hfl/releases)
2. Extract the binary to your PATH
3. Run `hfl --version` to verify

### From Source
```bash
git clone https://github.com/ahmaruff/hfl.git
cd hfl
go build -o hfl
```

### System Requirements
- No dependencies required
- Works on Windows, macOS, and Linux
- Minimum 10MB disk space

## Basic Usage

### Creating Your First Entry
```bash
hfl edit
```

This creates `hfl.md` in your current directory and opens today's entry in your configured editor.

### Editing Specific Dates
```bash
hfl edit 2025-08-20        # Specific date
hfl edit yesterday         # Yesterday's entry
hfl edit n+5              # 5 days from now
```

### Checking Your Journal
```bash
hfl check                  # Validate format
hfl status                 # Show sync status
```

## Date Parsing

HFL supports flexible date input:

### Relative Dates
```bash
hfl edit today             # Today
hfl edit yesterday         # Yesterday  
hfl edit tomorrow          # Tomorrow
```

### Offset Notation
```bash
hfl edit n+0               # Today (now + 0 days)
hfl edit n+1               # Tomorrow
hfl edit n-1               # Yesterday
hfl edit n+7               # Next week
hfl edit n-30              # 30 days ago
```

### Weekdays
```bash
hfl edit monday            # Next Monday
hfl edit friday            # Next Friday
hfl edit saturday          # Next Saturday
```

### Specific Dates
```bash
hfl edit 2025-08-20        # ISO format
hfl edit 2025-12-25        # Christmas 2025
```

## Configuration

### Configuration Hierarchy
HFL uses a layered configuration system (highest priority first):

1. **Environment variables** (e.g., `HFL_EDITOR`)
2. **Local config** (`./.hfl/config.json`)
3. **Global config** (`~/.hfl/config.json`)
4. **Built-in defaults**

### Managing Configuration
```bash
# Set values
hfl config set editor "code"
hfl config set editor "vim" --global

# Get values
hfl config get editor       # Specific key
hfl config get              # All settings

# List available keys
hfl config list
```

### Configuration Keys

| Key | Description | Example |
|-----|-------------|---------|
| `editor` | Text editor command | `"code"`, `"vim"`, `"nano"` |
| `conflict_strategy` | Sync conflict resolution | `"remote"`, `"local"` |
| `notion.api_token` | Notion integration token | `"ntn_xxx..."` |
| `notion.database_id` | Notion database ID | `"abc123..."` |

### Editor Configuration
```bash
# Method 1: Config command
hfl config set editor "code"

# Method 2: Environment variable
export HFL_EDITOR="vim"

# Method 3: System default
export EDITOR="nano"
```

**Editor Priority:** `HFL_EDITOR` → `EDITOR` → config file → default (`vi`/`notepad`)

## Notion Integration

### Setup Steps

#### 1. Create Notion Integration
1. Go to [https://www.notion.so/my-integrations](https://www.notion.so/my-integrations)
2. Click "New integration"
3. Name it "HFL" and select your workspace
4. Copy the "Internal Integration Token"

#### 2. Create Journal Database
Create a new database in Notion with these properties:
- **Date** (Date property)
- **HFL_Date** (Rich Text property)
- **Created** (Created time - optional)
- **Last Modified** (Last edited time - optional)
- **Word Count** (Number - optional)
- **Sync Status** (Select - optional)

#### 3. Share Database with Integration
1. Open your database in Notion
2. Click "Share" → "Add connections"
3. Select your HFL integration

#### 4. Configure HFL
```bash
# Get database ID from URL: https://notion.so/workspace/DatabaseName-DATABASE_ID?v=...
hfl config set notion.database_id "your-database-id"
hfl config set notion.api_token "ntn_your-token"
```

### Sync Commands
```bash
# Two-way sync (recommended)
hfl sync

# One-way sync
hfl sync --push             # Local → Notion
hfl sync --pull             # Notion → Local

# Preview changes
hfl sync --dry-run
```

### Conflict Resolution
Configure how conflicts are handled:
```bash
hfl config set conflict_strategy "remote"    # Notion wins
hfl config set conflict_strategy "local"     # Local wins
```

**Note:** `"merge"` strategy is planned for future releases.

## Commands Reference

### Core Commands

#### `hfl edit [DATE]`
Open journal entry for editing.
```bash
hfl edit                   # Today
hfl edit 2025-08-20       # Specific date
hfl edit n-5              # 5 days ago
```

#### `hfl check`
Validate journal format.
```bash
hfl check                  # Show warnings and errors
```
Exit codes: `0` (clean), `1` (errors), `2` (warnings)

#### `hfl status`
Show sync status.
```bash
hfl status                 # Local vs Notion comparison
```

### Export Commands

#### `hfl export json`
Export journal as JSON.
```bash
hfl export json            # To stdout
hfl export json -o backup.json
```

#### `hfl export csv`
Export as tab-separated CSV.
```bash
hfl export csv             # To stdout  
hfl export csv -o data.tsv
```

### Sync Commands

#### `hfl sync`
Synchronize with Notion.
```bash
hfl sync                   # Two-way sync
hfl sync --push           # Push to Notion
hfl sync --pull           # Pull from Notion  
hfl sync --dry-run        # Preview only
```

### Configuration Commands

#### `hfl config set`
Set configuration values.
```bash
hfl config set editor "code"
hfl config set editor "vim" --global
```

#### `hfl config get`
Get configuration values.
```bash
hfl config get editor     # Specific key
hfl config get            # All settings
```

#### `hfl config list`
List available configuration keys.
```bash
hfl config list
```

## File Format

### Structure
HFL uses a strict but simple Markdown format:

```markdown
# 2025-08-16
First paragraph of the entry.

Second paragraph with **formatting**.

# 2025-08-15
Another entry on a different date.
```

### Rules
- **Headings:** Must be `# YYYY-MM-DD` (H1 level only)
- **Body:** Free-form Markdown content
- **Order:** Entries sorted by date (newest first)  
- **Spacing:** Exactly one blank line between entries
- **Encoding:** UTF-8 without BOM
- **Line endings:** LF (`\n`)

### Valid Examples
```markdown
# 2025-08-16
Simple entry.

# 2025-08-15
Entry with *formatting* and **bold text**.

Multiple paragraphs are supported.

# 2025-08-14
- Lists work too
- Second item
```

### Invalid Examples
```markdown
## 2025-08-16           ❌ Wrong heading level
# Aug 16, 2025          ❌ Wrong date format  
# 2025-8-16             ❌ Missing leading zeros
# 2025-08-16            ❌ Duplicate date
Some content
# 2025-08-16            
More content
```

### Auto-formatting
HFL automatically formats your journal:
- Sorts entries by date (newest first)
- Normalizes spacing between entries
- Preserves body content exactly as written
- Removes trailing blank lines

## Troubleshooting

### Common Issues

#### Editor Not Opening
```bash
# Check current editor configuration
hfl config get editor

# Set explicitly
hfl config set editor "code"

# Test with environment variable
HFL_EDITOR="nano" hfl edit
```

#### Sync Failures
```bash
# Check configuration
hfl config get notion.api_token
hfl config get notion.database_id

# Test with dry run
hfl sync --dry-run

# Check file format
hfl check
```

#### Permission Errors
```bash
# Check file permissions
ls -la hfl.md

# Check directory permissions
ls -la .hfl/
```

<!-- ### Debug Mode -->
<!-- Enable verbose output by setting environment variable: -->
<!-- ```bash -->
<!-- HFL_DEBUG=1 hfl sync -->
<!-- ``` -->

### Getting Help
```bash
hfl help                   # General help
hfl edit --help           # Command-specific help
hfl config --help         # Config help
```

### Error Codes
- `0` - Success
- `1` - General error (file not found, format error, etc.)
- `2` - Warnings found (validation issues)

### Log Files
HFL doesn't create log files by default. All output goes to stdout/stderr.

<!-- ## Advanced Usage -->
<!---->
<!-- ### Batch Operations -->
<!-- ```bash -->
<!-- # Check multiple files -->
<!-- for file in *.md; do hfl check "$file"; done -->
<!---->
<!-- # Export with date range (future feature) -->
<!-- hfl export json --since 2025-01-01 --until 2025-12-31 -->
<!-- ``` -->

### Git Integration
```bash
# Initialize git repo
git init
echo ".hfl/" >> .gitignore  # Exclude sync state
git add hfl.md README.md
git commit -m "Initial journal"

# Daily commits
echo 'hfl check && git add hfl.md && git commit -m "Journal update"' > daily-backup.sh
```

### Scripting
```bash
#!/bin/bash
# Morning journal routine
hfl edit
hfl check
hfl sync --push
echo "Journal updated and synced!"
```

<!-- For more examples and community scripts, see the [examples directory](examples/) in the repository. -->
