# HFL (Homework-for-Life)

**A simple, powerful journaling tool that keeps your thoughts in sync.**

Write in your favorite editor, sync to Notion, access anywhere. One file, infinite possibilities.

## Why HFL?

🎯 **One file, everywhere** - Your entire journal in a single `hfl.md` file  
📝 **Write your way** - Use any editor you love (VS Code, Vim, Nano, or Notion)  
🔄 **Always in sync** - Seamless two-way sync with Notion  
📱 **Access anywhere** - Edit on desktop, phone, or web  
🚀 **Future-proof** - Plain Markdown that works with any tool  

## Quick Start

### 1. Install HFL
```bash
# Download from releases or build from source
git clone https://github.com/ahmaruff/hfl.git
cd hfl && go build -o hfl
```

### 2. Start writing
```bash
hfl edit                    # Open today's entry
hfl edit yesterday          # Yesterday's thoughts
hfl edit n+7               # Plan for next week
```

### 3. Sync with Notion (optional)
```bash
# Set up Notion integration
hfl config set notion.api_token "your-token"
hfl config set notion.database_id "your-db-id"

# Sync your journal
hfl sync                    # Two-way sync
hfl sync --push            # Local → Notion
hfl sync --pull            # Notion → Local
```

## Your Journal, Your Way

```markdown
# 2025-08-16
Had a breakthrough with the new project today.

The key insight was **simplicity over complexity**.

# 2025-08-15
Quiet morning. Read about minimalism.

Sometimes the best solutions are the simplest ones.
```

## What Makes HFL Special?

### 🎪 **Smart Date Parsing**
```bash
hfl edit today             # Today
hfl edit n-3              # 3 days ago  
hfl edit monday           # Next Monday
hfl edit 2025-12-25       # Specific date
```

### ⚡ **Instant Validation**
- Auto-formats on save
- Warns about issues
- Keeps your journal clean

### 🔧 **Flexible Configuration**
```bash
hfl config set editor "code"           # VS Code
hfl config set editor "vim" --global   # Vim everywhere
hfl config list                        # See all options
```

### 📊 **Rich Export Options**
```bash
hfl export json -o backup.json        # JSON backup
hfl export csv -o analytics.csv       # Data analysis
```

## Real-World Usage

**Morning pages** - `hfl edit` → write → auto-sync to Notion  
**Project logs** - Track daily progress with timestamps  
**Travel journal** - Write offline, sync when connected  
**Team updates** - Share via Notion, edit locally  

## What's Coming?

- 🎨 Rich markdown formatting in Notion
- 📊 Analytics and insights
- 🔗 Integration with more platforms
- 📱 Mobile companion app

## Installation Options

### Quick Install (Recommended)
Download from [releases](https://github.com/ahmaruff/hfl/releases) - binaries for Windows, macOS, and Linux.

### Build from Source
```bash
git clone https://github.com/ahmaruff/hfl.git
cd hfl
go build -o hfl
```

## Community

- 📖 **[Documentation](DOCUMENTATION.md)** - Complete guide
- 🐛 **[Issues](https://github.com/ahmaruff/hfl/issues)** - Bug reports & features
- 💬 **[Discussions](https://github.com/ahmaruff/hfl/discussions)** - Community support

## License

MIT License - your journal, your rules.

---

**Start your journey:** `hfl edit`

*Simple journaling that grows with you.*
