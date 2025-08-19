// Create internal/notion/client_test.go
package notion

import (
	"os"
	"testing"
)

func TestClient_QueryDatabase(t *testing.T) {
	token := os.Getenv("NOTION_API_TOKEN")
	dbID := os.Getenv("NOTION_DATABASE_ID")

	if token == "" || dbID == "" {
		t.Skip("Set NOTION_API_TOKEN and NOTION_DATABASE_ID to run this test")
	}

	client := NewClient(token)

	response, err := client.QueryDatabase(dbID, nil)
	if err != nil {
		t.Fatalf("QueryDatabase failed: %v", err)
	}

	t.Logf("Found %d pages", len(response.Results))
}
