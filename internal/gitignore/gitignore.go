package gitignore

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func EnsureHFLIgnored() error {
	const gitignoreFile = ".gitignore"
	const hflEntry = ".hfl"

	// Check if .gitignore exists
	if _, err := os.Stat(gitignoreFile); os.IsNotExist(err) {
		// create new file
		return createGitignore(gitignoreFile, hflEntry)

	}

	// Check if .hfl/ already ignored
	ignored, err := isAlreadyIgnored(gitignoreFile, hflEntry)
	if err != nil {
		return err
	}

	if !ignored {
		// Append .hfl/ to existing .gitignore
		return appendToGitignore(gitignoreFile, hflEntry)
	}

	return nil

}

func createGitignore(filename string, entry string) error {
	content := fmt.Sprintf("# HFL state and config\n%s\n", entry)
	return os.WriteFile(filename, []byte(content), 0644)
}

func isAlreadyIgnored(filename string, entry string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == entry || line == strings.TrimSuffix(entry, "/") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

func appendToGitignore(filename, entry string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add newline + comment + entry
	content := fmt.Sprintf("\n# HFL state and config\n%s\n", entry)
	_, err = file.WriteString(content)
	return err
}
