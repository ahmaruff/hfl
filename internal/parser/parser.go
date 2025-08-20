package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ParseFile(filename string) (*Journal, []string, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var journal Journal
	var warnings []string
	seenDates := make(map[string]int)

	headingRegex := regexp.MustCompile(`^# (\d{4})-(\d{2})-(\d{2})$`)

	var currentEntry *Entry
	var bodyLines []string
	var headerLines []string
	foundFirstEntry := false

	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			if !foundFirstEntry {
				foundFirstEntry = true

				if len(headerLines) > 0 {
					for len(headerLines) > 0 && strings.TrimSpace(headerLines[len(headerLines)-1]) == "" {
						headerLines = headerLines[:len(headerLines)-1]
					}
					if len(headerLines) > 0 {
						journal.Header = strings.Join(headerLines, "\n") + "\n"
					}
				}
			}

			date := matches[1] + "-" + matches[2] + "-" + matches[3] // Extract YYYY-MM-DD

			// Validate month and day
			month := matches[2]
			day := matches[3]

			monthInt, _ := strconv.Atoi(month) // No error check needed
			dayInt, _ := strconv.Atoi(day)

			if monthInt < 1 || monthInt > 12 || dayInt < 1 || dayInt > 31 {
				warnings = append(warnings, fmt.Sprintf("WARN invalid heading at line %d: \"%s\" (expected \"# YYYY-MM-DD\")", lineNum, line))
				continue // Skip this entry
			}

			if firstLine, exists := seenDates[date]; exists {
				warnings = append(warnings, fmt.Sprintf("WARN duplicate date %s at line %d (first seen at line %d); discarding duplicate", date, lineNum, firstLine))
				continue
			}

			seenDates[date] = lineNum

			if currentEntry != nil {
				// Remove trailing empty lines from bodyLines
				for len(bodyLines) > 0 && bodyLines[len(bodyLines)-1] == "" {
					bodyLines = bodyLines[:len(bodyLines)-1]
				}

				bodyContent := strings.Join(bodyLines, "\n")
				bodyContent = strings.TrimPrefix(bodyContent, "\n")
				currentEntry.Body = bodyContent

				journal.Entries = append(journal.Entries, *currentEntry)
			}

			currentEntry = &Entry{Date: date}

			// reset body
			bodyLines = []string{}

		} else {
			if !foundFirstEntry {
				headerLines = append(headerLines, line)
			} else if currentEntry != nil {
				bodyLines = append(bodyLines, line)
			} else {
				warnings = append(warnings, fmt.Sprintf("WARN text before first heading at line %d", lineNum))
			}
		}
	}

	// Don't forget the last entry!
	if currentEntry != nil {
		// Remove trailing empty lines from bodyLines
		for len(bodyLines) > 0 && bodyLines[len(bodyLines)-1] == "" {
			bodyLines = bodyLines[:len(bodyLines)-1]
		}

		bodyContent := strings.Join(bodyLines, "\n")
		bodyContent = strings.TrimPrefix(bodyContent, "\n")
		currentEntry.Body = bodyContent

		journal.Entries = append(journal.Entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, warnings, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return &journal, warnings, nil
}
