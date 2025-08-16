# HFL (Homework‑for‑Life) — MVP Specification v1.0

Version: 1.0
Status: Final (MVP)
Date: 2025‑08‑16

This document is tool‑agnostic and normative. It defines a minimal, deterministic format + behavior so multiple implementations (Go, Rust, Python, etc.) interoperate reliably.

---

## 0) Scope & Goals (normative)

### 0.1 Goals

Simplicity: one human‑editable Markdown file (hfl.md).  
Determinism: canonical formatting so parse → rewrite → parse is stable.  
Interoperability: clear grammar and file schemas so different tools produce identical results.  
Offline‑first: all core features work without network; Notion sync is optional.  
Safety: diagnostics go to console; state files remain clean of logs.  

### 0.2 Non‑Goals (MVP)

Tags/metadata inside Markdown entries.  
Multiple entries per date.  
Hard delete semantics across local/remote (only create/update in MVP).  
Full calendrical correctness beyond month/day ranges (leap‑year checks optional).  

---

## 1) Terminology (normative)

**Entry** : One day’s record = a Heading line + a Body block.  
**Heading**: Markdown H1 line in the exact form # YYYY-MM-DD.  
**Body** : All lines after a heading up to (but not including) the next heading or EOF.  
**Journal file** : hfl.md in the project directory.  
**Project directory** : Folder containing `hfl.md` and `.hfl/`.  
**Global config** : ~/.hfl/config.json.  
**Local config** : ./.hfl/config.json (overrides global, key by key).  
**MUST/SHOULD/MAY** : As defined in RFC 2119.  

---

## 2) File Layout (normative)

```
project/
├─ hfl.md
└─ .hfl/
   ├─ config.json     # user-settable settings (local; overrides global)
   └─ state.json      # tool-managed mapping/state (no diagnostics)
```

---

## 3) Encoding & Newlines (normative)

**Encoding** : UTF‑8 without BOM (MUST).  
**Read** : Tools SHOULD accept LF or CRLF.  
**Write** : Tools MUST output LF (\n).  
**EOF** : File SHOULD end with a single newline. The file MUST NOT end with an extra blank line paragraph (see §6.2).  

---

## 4) hfl.md Format (normative)

### 4.1 Grammar (EBNF)

```ebnf
File          = Entry , { BlankLine , Entry } , [ Newline ] ;

Entry         = Heading , Newline , Body ;

Heading       = "#" , " " , Date ;

Date          = Year , "-" , Month , "-" , Day ;

Year          = Digit , Digit , Digit , Digit ;
Month         = Digit , Digit ;     (* 01..12 validated separately *)
Day           = Digit , Digit ;     (* 01..31 validated separately *)

Body          = { BodyLine } ;

BodyLine      = { any-char-except-LF } , Newline
              | Newline ;

BlankLine     = Newline ;

Newline       = "\n" ;
Digit         = "0"…"9" ;
```

### 4.2 Heading Regex

Exact match: `^# (\d{4})-(\d{2})-(\d{2})$`  
Tools MUST validate month in 01..12 and day in 01..31. (Leap‑day validity MAY be enforced; see §0.2.)  

### 4.3 Constraints

**Uniqueness** : Exactly one entry per date within the file.  
**Heading level** : MUST be H1 (#). Other levels are invalid.  
**Order (canonical)** : Descending by date (newest first). Writers MUST produce canonical order; readers MAY accept any order.  
**Body**: Free‑form Markdown. Writers MUST NOT reflow/alter body content.  

### 4.4 Examples

**Valid**

```md
# 2025-08-16
First paragraph.

Second paragraph with *markdown*.

# 2025-08-15
One-liner.
```

**Invalid**

```md
## 2025-08-16 (wrong level)

# Aug 16, 2025 (wrong format)
duplicate


# 2025-08-16 (non‑unique)
two blank lines between entries (spacing violation; normalized on write)
```

---

# 5) Parser Behavior (normative)

MUST produce a list of entries:
```
[
  { "date": "YYYY-MM-DD", "body": "verbatim-with-newlines" }
]
```

MUST collect violations (see §7) and print warnings to console.  
MUST be tolerant on read (best‑effort) but MUST drop entries that fail Heading Regex or Date Uniqueness.  
Dropped entries MUST be reported.

---

## 6) Canonical Rewrite Rules (normative)

### 6.1 Ordering

Sort entries by date descending (ISO lexicographic ordering suffices).

### 6.2 Spacing

Between entries: exactly one blank line (i.e., one `\n` line separating the last body line and the next `#`).  
EOF: file SHOULD end with a single trailing `\n`. Do not add an additional empty paragraph at end  
(no two consecutive `\n` after the final body line).  

### 6.3 Heading Form

Always write headings exactly as # YYYY-MM-DD.

### 6.4 Body Preservation

Write body byte‑for‑byte as parsed (whitespace and Markdown untouched).  
Implication: Parse → Canonical rewrite → Parse is idempotent modulo ordering/spacing normalization.

---

## 7) Diagnostics & Invalid Cases (normative)

All diagnostics are console‑only; implementations MUST NOT write diagnostics into state.json.

A) Invalid Heading Format  
Action: drop entry;  
warn: `WARN invalid heading at line N: "# Aug 16, 2025" (expected "# YYYY-MM-DD")`  

B) Duplicate Date  
Action: keep first occurrence, drop subsequent;  
warn: `WARN duplicate date 2025-08-16 at line N (first seen at line M); discarding duplicate`  

C) Multiple Blank Lines Between Entries  
Action on write: normalize to one blank line.  
Implementations MAY log: `INFO normalized blank lines between 2025-08-14 and 2025-08-13`  

D) Text Outside Any Entry  
(e.g., preface before the first heading)  
Action: ignore;  
warn.

E) Calendar Validity  
Enforce month 01..12 and day 01..31.  
Leap‑day validity MAY be enforced; if enforced and invalid → treat as Invalid Heading.  

---

## 8) Config (.hfl/config.json) — Schema & Resolution (normative)

### 8.1 JSON Schema (Draft 2020‑12)

```
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "editor": { "type": "string" },
    "conflict_strategy": {
      "type": "string",
      "enum": ["remote", "local", "merge"],
      "default": "remote"
    },
    "notion": {
      "type": "object",
      "properties": {
        "api_token": { "type": "string" },
        "database_id": { "type": "string" }
      },
      "required": [],
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}
```

### 8.2 Resolution Priority & Defaults

Effective config = merge of, in ascending precedence:

- Global `~/.hfl/config.json`  
- Local `./.hfl/config.json`  
- Environment overrides (highest):  
  - `HFL_EDITOR` → editor  
  - `HFL_NOTION_TOKEN` → notion.api_token  
  - `HFL_NOTION_DATABASE` → notion.database_id  

Editor default resolution:  
`config.editor` → `env($HFL_EDITOR)` → `env($EDITOR)` → `"vi"`.

---

## 9) State (.hfl/state.json) — Schema (normative)

### 9.1 Purpose

Tool‑managed metadata to support sync. No diagnostics/history.

### 9.2 JSON Schema

```
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "last_synced": { "type": "string", "format": "date-time" },
    "entries": {
      "type": "object",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "notion_id": { "type": ["string", "null"] },
          "hash": { "type": ["string", "null"] },
          "last_remote_edit": { "type": ["string", "null"], "format": "date-time" },
          "last_local_sync": { "type": ["string", "null"], "format": "date-time" }
        },
        "required": [],
        "additionalProperties": false
      }
    }
  },
  "required": ["entries"],
  "additionalProperties": false
}
```

Keys of entries are date strings YYYY-MM-DD.

---

## 10) CLI Contract (normative)

### 10.1 `edit [DATE]`

Opens hfl.md in the resolved editor.  
If DATE omitted → use today (system local timezone).  
If entry exists → MAY position cursor at its body; otherwise create canonical heading + empty body.  
Exit codes: 0 success; 1 IO/format error.

### 10.2 `check`

Parses hfl.md; prints warnings per §7.  
No file modifications.  
Exit codes: 0 no warnings; 2 warnings present; 1 fatal error (e.g., unreadable file).

### 10.3 `sync`

Preconditions: effective config contains `notion.api_token` and `notion.database_id`.  
Behavior (MVP):  
For each local date:  
- If no notion_id → create remote page; store notion_id, hash, timestamps.  
- If hash differs and remote last_remote_edit is older than local → update remote.  
- If remote is newer than last_local_sync → pull and overwrite local body (unless conflictStrategy dictates otherwise).  
- For remote pages absent locally → add new local entry.  

Conflict Strategy:  
`remote` (default): remote overwrites local.  
`local` : local overwrites remote.  
`merge`: Not supported in MVP. Implementations MUST treat this as a fatal error and exit.

After completion: update `state.json`; rewrite hfl.md canonically (see §6).  
Exit codes: 0 success; 1 error.

### 10.4 `export json`

Emit JSON array of `{ date, body }` to STDOUT or file (flag‑controlled).  
Exit code: 0 success; 1 error.

---

## 11) Notion Interop (informative but consistent)

Recommended database schema:  
- Date (Date property) — required; one page per date.
- Note (Rich Text) — full body.
- HFL_Date (Plain Text) — exact YYYY-MM-DD for lookup/uniqueness.

Timestamps:  
- Remote freshness: Notion last_edited_time.
- Local freshness: derived from stored hash and/or last_local_sync.

--- 

## 12) Hashing (recommended)

`state.json.entries[date].hash` = SHA‑256 over the body bytes (UTF‑8).  
The hash MUST be calculated over the raw body content as it exists in the hfl.md file,  
exactly as parsed, without any normalization or reformatting.

---

## 13) Security (normative)

`config.json` may contain secrets (e.g., notion.api_token). 
Implementations MUST advise users to keep .hfl/ out of VCS (e.g., via .gitignore).  
Implementations MUST NOT print secrets in logs or diagnostics.

---

## 14) Determinism (normative)

Given identical `hfl.md` and `state.json`, two compliant implementations MUST produce byte‑identical canonical hfl.md after a no‑op sync or format pass.

---

## 15) Canonicalization Test Vectors (normative)

### 15.1 Input (messy)

```md
# 2025-08-15
Something.

# Aug 16, 2025
Bad heading.

# 2025-08-16
First valid.

# 2025-08-16
Duplicate.

# 2025-08-14
Older entry.


# 2025-08-13
With extra blank line.
```


### 15.2 Output (canonical)

```
# 2025-08-16
First valid.

# 2025-08-15
Something.

# 2025-08-14
Older entry.

# 2025-08-13
With extra blank line.
```

### 15.3 Console Warnings

```
WARN invalid heading at line 5: "# Aug 16, 2025" (expected "# YYYY-MM-DD")
```

```
WARN duplicate date 2025-08-16 at line 9 (first seen at line 7); discarding duplicate
```
