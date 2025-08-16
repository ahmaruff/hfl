# HFL (Homework-for-Life)
A simple, deterministic journaling tool that captures your daily life in a single Markdown file.

## What is HFL?

HFL helps you build a daily writing habit by keeping all entries in one plain-text file (hfl.md).
Each entry is stored under a date heading in standard Markdown, making it easy to read, edit, and version-control:

```markdown
# 2025-08-16
Had a productive day working on the new project.

Made good progress on the parser implementation.

# 2025-08-15
Quiet weekend. Read a good book.
```

No databases, no proprietary formats—just your thoughts in Markdown.

## Features

- **One file, one journal** : Everything lives in hfl.md—simple to back up, sync, or put under Git.
- **Plain Markdown** : Portable, future-proof, and readable in any editor or Markdown viewer.
- **Deterministic format** : Entries follow a predictable structure, so they’re easy to parse, search, and extend.

- **Offline-first** : Works without internet or setup. Just open your editor and write.

Future-ready
Extensible by design—planned features (like Notion sync) won’t break the core simplicity.
- **Single file**: Everything in one `hfl.md` file
- **Human-readable**: Standard Markdown format
- **Deterministic**: Consistent formatting across tools
- **Offline-first**: Works without internet
- **Extensible**: Built for future Notion sync (coming soon)

## Installation

### From Source

```bash
git clone https://github.com/ahmaruff/hfl.git
cd hfl
go build -o hfl
```

### Direct Download

Download the latest release from the [releases page](https://github.com/ahmaruff/hfl/releases).

## Usage

### Create or edit today's entry
```bash
hfl edit
```

### Edit a specific date
```bash
hfl edit                   # Today
hfl edit yesterday         # Yesterday
hfl edit tomorrow          # Tomorrow
hfl edit n+0               # Today
hfl edit n+1               # Tomorrow  
hfl edit n-1               # Yesterday
hfl edit n-30              # 30 days ago
hfl edit monday            # Next Monday
hfl edit 2025-08-20        # Specific date
```


### Check file for formatting issues
```bash
hfl check
```

### Export your journal
```bash
# Export as JSON
hfl export json
hfl export json -o backup.json

# Export as CSV (tab-separated)
hfl export csv
hfl export csv -o data.tsv
```

### Configure HFL
```bash
# Set your preferred editor
hfl config set editor "code"
hfl config set editor "nvim" --global

# View current configuration
hfl config get
hfl config get editor

# List available configuration keys
hfl config list
```

## File Format

HFL uses a strict but simple format:

- **Headings**: Must be `# YYYY-MM-DD` (H1 level only)
- **Body**: Free-form Markdown content
- **Order**: Entries sorted by date (newest first)
- **Spacing**: Exactly one blank line between entries

### Valid Example
```markdown
# 2025-08-16
First paragraph.

Second paragraph with *markdown*.

# 2025-08-15
Single line entry.
```

### Invalid Examples
```markdown
## 2025-08-16  ❌ Wrong heading level
# Aug 16, 2025  ❌ Wrong date format
```

## Editor Configuration

HFL supports flexible configuration through multiple methods:

### Configuration Sources (in order of precedence)
1. **Environment variables** (highest priority)
2. **Local config** (`./.hfl/config.json`)
3. **Global config** (`~/.hfl/config.json`)
4. **Built-in defaults** (lowest priority)

### Editor Configuration
```bash
# Method 1: Using config command (recommended)
hfl config set editor "code"           # Local config
hfl config set editor "vim" --global   # Global config

# Method 2: Environment variables
export HFL_EDITOR="code"  # VS Code
export HFL_EDITOR="vim"   # Vim
export HFL_EDITOR="nano"  # Nano
export EDITOR="emacs"     # System default
```

### Available Configuration Keys
- `editor` - Your preferred text editor
- `conflict_strategy` - Sync conflict resolution (remote, local, merge)
- `notion.api_token` - Notion API token for sync (coming soon)
- `notion.database_id` - Notion database ID for sync (coming soon)

## Commands

| Command | Description |
|---------|-------------|
| `hfl edit [DATE]` | Create/edit journal entry |
| `hfl check` | Validate hfl.md format |
| `hfl export json [-o FILE]` | Export as JSON |
| `hfl export csv [-o FILE]` | Export as tab-separated CSV |
| `hfl config set <key> <value>` | Set configuration value |
| `hfl config get [key]` | Get configuration value(s) |
| `hfl config list` | List available configuration keys |

## Exit Codes

- `0`: Success
- `1`: Error (file not found, format error)
- `2`: Warnings found (check command)

## Validation & Warnings

HFL validates your entries and provides helpful warnings:

```bash
$ hfl check
WARN invalid heading at line 5: "# Aug 16, 2025" (expected "# YYYY-MM-DD")
WARN duplicate date 2025-08-16 at line 9 (first seen at line 7); discarding duplicate
```

## Upcoming Features

- **Notion sync**: Two-way sync with Notion databases

## Technical Details

For the complete technical specification, see [SPEC.md](SPEC.md).

Key points:
- UTF-8 encoding without BOM
- LF line endings (`\n`)
- Deterministic parsing and formatting
- Tolerant reader, strict writer

## Contributing

1. Check the [specification](SPEC.md) for implementation details
2. Fork the repository
3. Create a feature branch
4. Add tests for new functionality
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Why HFL?

- **Simple**: One file, standard Markdown
- **Reliable**: Deterministic format, extensive validation
- **Future-proof**: Designed for tool interoperability
- **Yours**: Local files, no vendor lock-in

Start journaling today with `hfl edit`!
