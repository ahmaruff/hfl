// sync.go
package notion

import (
	"fmt"
	"strings"
	"time"

	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/state"
)

type SyncService struct {
	client     *Client
	databaseID string
}

func NewSyncService(token, databaseID string) *SyncService {
	return &SyncService{
		client:     NewClient(token),
		databaseID: databaseID,
	}
}

func (s *SyncService) ValidateAndSetupDatabase() error {
	db, err := s.client.GetDatabase(s.databaseID)
	if err != nil {
		return fmt.Errorf("failed to fetch database: %w", err)
	}

	props, ok := db["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected schema format")
	}

	required := map[string]map[string]interface{}{
		"Date":       {"date": map[string]interface{}{}},
		"HFL_Date":   {"rich_text": map[string]interface{}{}},
		"Word Count": {"number": map[string]interface{}{}},
		"Sync Status": {
			"select": map[string]interface{}{
				"options": []map[string]interface{}{
					{"name": "Pending", "color": "blue"},
					{"name": "Conflict", "color": "red"},
					{"name": "Modified", "color": "yellow"},
					{"name": "Synced", "color": "green"},
				},
			},
		},
	}

	missing := make(map[string]interface{})
	for key, schema := range required {
		if _, exists := props[key]; !exists {
			missing[key] = schema
		}
	}

	if len(missing) > 0 {
		fmt.Printf("Adding missing properties: %v\n", missing)
		if err := s.client.UpdateDatabase(s.databaseID, missing); err != nil {
			return fmt.Errorf("failed to update database schema: %w", err)
		}
	} else {
		fmt.Println("Database schema is valid")
	}

	return nil
}

// SyncToNotion pushes local changes to Notion
func (s *SyncService) SyncToNotion(journal *parser.Journal, state *state.State) error {
	for _, entry := range journal.Entries {
		entryState, exists := state.GetEntry(entry.Date)

		if !exists || entryState.NotionID == "" {
			if err := s.createEntry(entry, state); err != nil {
				return fmt.Errorf("failed to create entry %s: %w", entry.Date, err)
			}
		} else if state.HasChanged(entry.Date, entry.Body) {
			if err := s.updateEntry(entry, entryState, state); err != nil {
				return fmt.Errorf("failed to update entry %s: %w", entry.Date, err)
			}
		}
	}

	// Update last sync time
	state.SetLastSynced(time.Now().Format(time.RFC3339))
	return state.Save()
}

// SyncFromNotion pulls remote changes from Notion
func (s *SyncService) SyncFromNotion(journal *parser.Journal, state *state.State) error {
	// Query all entries from Notion
	response, err := s.client.QueryDatabase(s.databaseID, nil)
	if err != nil {
		return fmt.Errorf("failed to query database: %w", err)
	}

	fmt.Printf("Found %d pages in Notion database\n", len(response.Results))

	for _, page := range response.Results {
		date := s.extractDate(page)
		if date == "" {
			continue
		}

		// Get content from page blocks only
		content, err := s.getPageContent(page.ID)
		if err != nil {
			fmt.Printf("Warning: failed to get content for %s: %v\n", date, err)
			continue
		}

		// Check if we need to update local entry
		entryState, exists := state.GetEntry(date)
		if !exists || s.shouldUpdateLocal(page, entryState) {
			if err := s.updateLocalEntry(journal, date, content, page, state); err != nil {
				return fmt.Errorf("failed to update local entry %s: %w", date, err)
			}
		}
	}

	fmt.Printf("Sync from Notion done\n")
	return nil
}

// extractNoteProperty: HAPUS atau RENAME jika tidak dipakai
// (Tidak dipakai di kode, jadi bisa dihapus)
// func (s *SyncService) extractNoteProperty(page Page) string { ... }

func (s *SyncService) createEntry(entry parser.Entry, state *state.State) error {
	wordCount := float64(len(strings.Fields(entry.Body)))

	properties := Properties{
		"Date":        NewDateProperty(entry.Date),
		"HFL_Date":    NewPlainRichTextProperty(entry.Date),
		"Word Count":  NewNumberProperty(wordCount),
		"Sync Status": NewSelectProperty("Synced"),
	}

	blocks := MarkdownToBlocks(entry.Body)

	page, err := s.client.CreatePage(s.databaseID, properties, blocks)
	if err != nil {
		return err
	}

	// Update state
	state.SetNotionID(entry.Date, page.ID)
	state.UpdateEntry(entry.Date, entry.Body)

	return nil
}

func (s *SyncService) updateEntry(entry parser.Entry, entryState state.EntryState, state *state.State) error {
	if entryState.NotionID == "" {
		return fmt.Errorf("no Notion ID for entry %s", entry.Date)
	}

	// Update metadata properties
	wordCount := float64(len(strings.Fields(entry.Body)))
	properties := Properties{
		"Word Count":  NewNumberProperty(wordCount),
		"Sync Status": NewSelectProperty("Synced"),
	}

	_, err := s.client.UpdatePage(entryState.NotionID, properties)
	if err != nil {
		return err
	}

	// Update page content
	blocks := MarkdownToBlocks(entry.Body)
	if err := s.client.UpdateBlockChildren(entryState.NotionID, blocks); err != nil {
		return err
	}

	// Update state
	state.UpdateEntry(entry.Date, entry.Body)

	return nil
}

// getPageContent: FIXED — baca semua blok dengan benar
func (s *SyncService) getPageContent(pageID string) (string, error) {
	blocks, err := s.client.GetBlockChildren(pageID)
	if err != nil {
		return "", fmt.Errorf("failed to get block children: %w", err)
	}

	var content strings.Builder

	for i, block := range blocks.Results {
		var textObjects []TextObject

		switch block.Type {
		case "paragraph":
			if block.Paragraph != nil {
				// Loop semua RichText, lalu ambil TextObject-nya
				for _, rt := range block.Paragraph.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		case "heading_1":
			if block.Heading1 != nil {
				for _, rt := range block.Heading1.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		case "heading_2":
			if block.Heading2 != nil {
				for _, rt := range block.Heading2.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		case "heading_3":
			if block.Heading3 != nil {
				for _, rt := range block.Heading3.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		case "bulleted_list_item":
			if block.BulletedListItem != nil {
				for _, rt := range block.BulletedListItem.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		case "numbered_list_item":
			if block.NumberedListItem != nil {
				for _, rt := range block.NumberedListItem.RichText {
					textObjects = append(textObjects, rt)
				}
			}
		default:
			fmt.Printf("Unsupported block type: %s (skipping)\n", block.Type)
			continue
		}

		// Ekstrak teks dari textObjects
		for _, obj := range textObjects {
			if obj.Text != nil {
				content.WriteString(obj.Text.Content)
			}
		}

		// Tambahkan spasi antar blok
		if i < len(blocks.Results)-1 {
			switch block.Type {
			case "paragraph", "heading_1", "heading_2", "heading_3":
				content.WriteString("\n\n")
			case "bulleted_list_item", "numbered_list_item":
				content.WriteString("\n")
			}
		}
	}

	finalContent := strings.TrimSpace(content.String())
	return finalContent, nil
}

// extractDate: FIXED — akses HFL_Date dengan benar
func (s *SyncService) extractDate(page Page) string {
	// Helper: cek format tanggal
	isValidDate := func(s string) bool {
		_, err := time.Parse("2006-01-02", s)
		return err == nil
	}

	// Try HFL_Date first (more reliable)
	if prop, ok := page.Properties["HFL_Date"]; ok && len(prop.RichText) > 0 {
		textObj := prop.RichText[0]
		if textObj.Text != nil {
			content := strings.TrimSpace(textObj.Text.Content)
			if isValidDate(content) {
				return content
			}
		}
	}

	// Fallback to Date property
	if dateProp, ok := page.Properties["Date"]; ok && dateProp.Date != nil {
		dateStr := strings.Split(dateProp.Date.Start, "T")[0]
		if isValidDate(dateStr) {
			return dateStr
		}
	}

	return ""
}

func (s *SyncService) shouldUpdateLocal(page Page, entryState state.EntryState) bool {
	if entryState.LastLocalSync == "" {
		return true
	}

	localTime, err := time.Parse(time.RFC3339, entryState.LastLocalSync)
	if err != nil {
		return true
	}

	return page.LastEditedTime.After(localTime)
}

func (s *SyncService) updateLocalEntry(journal *parser.Journal, date, content string, page Page, state *state.State) error {
	found := false
	for i, entry := range journal.Entries {
		if entry.Date == date {
			journal.Entries[i].Body = content
			found = true
			break
		}
	}

	if !found {
		newEntry := parser.Entry{
			Date: date,
			Body: content,
		}
		journal.Entries = append(journal.Entries, newEntry)
	}

	// Update Notion metadata
	wordCount := float64(len(strings.Fields(content)))
	properties := Properties{
		"Word Count":  NewNumberProperty(wordCount),
		"Sync Status": NewSelectProperty("Synced"),
	}

	if _, err := s.client.UpdatePage(page.ID, properties); err != nil {
		fmt.Printf("Warning: failed to update metadata for %s: %v\n", date, err)
	}

	// Update state
	state.SetNotionID(date, page.ID)
	state.UpdateEntry(date, content)

	return nil
}
