// types.go
package notion

import (
	"regexp"
	"strings"
	"time"
)

// Core response types
type QueryResponse struct {
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

type Page struct {
	ID             string     `json:"id"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	Properties     Properties `json:"properties"`
	URL            string     `json:"url"`
}

type BlockListResponse struct {
	Results []Block `json:"results"`
	HasMore bool    `json:"has_more"`
}

// Properties for database pages
type Properties map[string]Property

type Property struct {
	Type     string       `json:"type"`
	Date     *DateProp    `json:"date,omitempty"`
	RichText []TextObject `json:"rich_text,omitempty"`
	Number   *float64     `json:"number,omitempty"`
	Select   *Select      `json:"select,omitempty"`
}

type DateProp struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

type Select struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type TextObject struct {
	Type        string       `json:"type"`
	Text        *Text        `json:"text,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

type Link struct {
	URL string `json:"url"`
}

type Text struct {
	Content string `json:"content"`
	Link    *Link  `json:"link,omitempty"`
}

type Annotations struct {
	Bold   bool   `json:"bold"`
	Italic bool   `json:"italic"`
	Code   bool   `json:"code"`
	Color  string `json:"color"`
}

// Block types for page content
type Block struct {
	ID               string          `json:"id,omitempty"`
	Type             string          `json:"type"`
	Paragraph        *ParagraphBlock `json:"paragraph,omitempty"`
	Heading1         *HeadingBlock   `json:"heading_1,omitempty"`
	Heading2         *HeadingBlock   `json:"heading_2,omitempty"`
	Heading3         *HeadingBlock   `json:"heading_3,omitempty"`
	BulletedListItem *ListItemBlock  `json:"bulleted_list_item,omitempty"`
	NumberedListItem *ListItemBlock  `json:"numbered_list_item,omitempty"`
	ToDo             *ToDoBlock      `json:"to_do,omitempty"`
	Quote            *QuoteBlock     `json:"quote,omitempty"`
	// Tambahkan tipe lain jika perlu
}

// Block content types
type ParagraphBlock struct {
	RichText []TextObject `json:"rich_text"`
}

type HeadingBlock struct {
	RichText []TextObject `json:"rich_text"`
	Color    string       `json:"color,omitempty"`
	IsToggle bool         `json:"is_toggle,omitempty"`
}

type ListItemBlock struct {
	RichText []TextObject `json:"rich_text"`
	Color    string       `json:"color,omitempty"`
	Children []Block      `json:"children,omitempty"` // untuk nested list
}

type ToDoBlock struct {
	RichText []TextObject `json:"rich_text"`
	Checked  bool         `json:"checked"`
	Color    string       `json:"color,omitempty"`
}

type QuoteBlock struct {
	RichText []TextObject `json:"rich_text"`
	Color    string       `json:"color,omitempty"`
	Citation string       `json:"citation,omitempty"`
}

// ParseTextObjects parses a string with simple Markdown formatting into []TextObject
func ParseTextObjects(text string) []TextObject {
	var objects []TextObject

	// Regex untuk [text](url)
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	parts := re.FindAllStringSubmatchIndex(text, -1)

	if len(parts) == 0 {
		// Tidak ada link → parsing biasa
		return []TextObject{
			{
				Type: "text",
				Text: &Text{
					Content: text,
				},
				Annotations: &Annotations{
					Bold:   false,
					Italic: false,
					Code:   false,
					Color:  "default",
				},
			},
		}
	}

	lastEnd := 0
	for _, match := range parts {
		start, end := match[0], match[1]
		linkTextStart, linkTextEnd := match[2], match[3]
		urlStart, urlEnd := match[4], match[5]

		// Teks sebelum link
		if start > lastEnd {
			plain := text[lastEnd:start]
			objects = append(objects, TextObject{
				Type: "text",
				Text: &Text{
					Content: plain,
				},
				Annotations: &Annotations{
					Bold:   strings.Contains(plain, "**"),
					Italic: strings.Contains(plain, "*"),
					Code:   strings.HasPrefix(plain, "`") && strings.HasSuffix(plain, "`"),
					Color:  "default",
				},
			})
		}

		// Bagian link
		linkText := text[linkTextStart:linkTextEnd]
		url := strings.TrimSpace(text[urlStart:urlEnd])

		objects = append(objects, TextObject{
			Type: "text",
			Text: &Text{
				Content: linkText,
				Link: &Link{
					URL: url,
				},
			},
			Annotations: &Annotations{
				Bold:   false,
				Italic: false,
				Code:   false,
				Color:  "default",
			},
		})

		lastEnd = end
	}

	// Setelah link terakhir
	if lastEnd < len(text) {
		after := text[lastEnd:]
		objects = append(objects, TextObject{
			Type: "text",
			Text: &Text{
				Content: after,
			},
			Annotations: &Annotations{
				Bold:   false,
				Italic: false,
				Code:   false,
				Color:  "default",
			},
		})
	}

	return objects
}

// parseInlineFormatting handles **bold**, *italic*, `code`
func parseInlineFormatting(text string) []TextObject {
	// Reuse your existing logic from ParseTextObjects (without link support)
	// Or keep the old ParseTextObjects but without link handling
	// We’ll extract just bold/italic/code
	return []TextObject{
		{
			Type: "text",
			Text: &Text{Content: text},
			Annotations: &Annotations{
				Bold:   strings.HasPrefix(text, "**") && strings.HasSuffix(text, "**"),
				Italic: strings.HasPrefix(text, "*") && strings.HasSuffix(text, "*"),
				Code:   strings.HasPrefix(text, "`") && strings.HasSuffix(text, "`"),
				Color:  "default",
			},
		},
	}
	// For full version, keep your old formatting parser (we simplified here)
}

// plainTextObject creates a TextObject with default annotations
func plainTextObject(content string) TextObject {
	return TextObject{
		Type: "text",
		Text: &Text{Content: content},
		Annotations: &Annotations{
			Bold:   false,
			Italic: false,
			Code:   false,
			Color:  "default",
		},
	}
}

// indexNotEscaped finds the first unescaped occurrence of substr in s
func indexNotEscaped(s, substr string) int {
	i := strings.Index(s, substr)
	for i > 0 && s[i-1] == '\\' {
		rest := s[i+len(substr):]
		next := strings.Index(rest, substr)
		if next == -1 {
			return -1
		}
		i += next + len(substr) + i + len(substr) - i // Perbaikan logika skip
		// Lebih aman: kita ubah jadi loop sederhana
	}
	return i
}

// Helper functions for creating blocks
func NewHeadingBlock(text string, level int) Block {
	textObjects := ParseTextObjects(text)

	switch level {
	case 1:
		return Block{
			Type: "heading_1",
			Heading1: &HeadingBlock{
				RichText: textObjects,
			},
		}
	case 2:
		return Block{
			Type: "heading_2",
			Heading2: &HeadingBlock{
				RichText: textObjects,
			},
		}
	case 3:
		return Block{
			Type: "heading_3",
			Heading3: &HeadingBlock{
				RichText: textObjects,
			},
		}
	default:
		return NewParagraphBlock(text)
	}
}

func NewListItemBlock(text string) Block {
	return Block{
		Type: "bulleted_list_item",
		BulletedListItem: &ListItemBlock{
			RichText: ParseTextObjects(text),
		},
	}
}

func NewNumberedListItemBlock(text string) Block {
	return Block{
		Type: "numbered_list_item",
		NumberedListItem: &ListItemBlock{
			RichText: ParseTextObjects(text),
		},
	}
}

func NewParagraphBlock(text string) Block {
	return Block{
		Type: "paragraph",
		Paragraph: &ParagraphBlock{
			RichText: ParseTextObjects(text),
		},
	}
}

// Helper functions for creating properties
func NewDateProperty(date string) Property {
	return Property{
		Type: "date",
		Date: &DateProp{Start: date},
	}
}

func NewRichTextProperty(text string) Property {
	return Property{
		Type:     "rich_text",
		RichText: ParseTextObjects(text),
	}
}

func NewPlainRichTextProperty(text string) Property {
	return Property{
		Type: "rich_text",
		RichText: []TextObject{
			{
				Type: "text",
				Text: &Text{
					Content: text,
				},
				Annotations: &Annotations{
					Bold:   false,
					Italic: false,
					Code:   false,
					Color:  "default",
				},
			},
		},
	}
}

func NewNumberProperty(num float64) Property {
	return Property{
		Type:   "number",
		Number: &num,
	}
}

func NewSelectProperty(value string) Property {
	return Property{
		Type: "select",
		Select: &Select{
			Name: value,
		},
	}
}

// MarkdownToBlocks converts simple Markdown to Notion blocks
func MarkdownToBlocks(markdown string) []Block {
	lines := strings.Split(markdown, "\n")
	blocks := []Block{}
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		switch {
		case strings.HasPrefix(line, "# "):
			blocks = append(blocks, NewHeadingBlock(line[2:], 1))
		case strings.HasPrefix(line, "## "):
			blocks = append(blocks, NewHeadingBlock(line[3:], 2))
		case strings.HasPrefix(line, "### "):
			blocks = append(blocks, NewHeadingBlock(line[4:], 3))
		case strings.HasPrefix(line, "- "):
			blocks = append(blocks, NewListItemBlock(line[2:]))
		case strings.HasPrefix(line, "1. "):
			blocks = append(blocks, NewNumberedListItemBlock(strings.TrimPrefix(line, "1. ")))
		case strings.HasPrefix(line, "```"):
			// TODO: handle code block
		default:
			blocks = append(blocks, NewParagraphBlock(line))
		}
		i++
	}

	return blocks
}

// splitParagraphs splits text into paragraphs (separated by blank lines)
func splitParagraphs(text string) []string {
	lines := strings.Split(text, "\n")
	var paragraphs []string
	var current string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if current != "" {
				paragraphs = append(paragraphs, strings.TrimSpace(current))
				current = ""
			}
		} else {
			if current != "" {
				current += "\n"
			}
			current += line
		}
	}

	if current != "" {
		paragraphs = append(paragraphs, strings.TrimSpace(current))
	}

	return paragraphs
}
