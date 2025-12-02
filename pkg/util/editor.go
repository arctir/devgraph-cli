package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// OpenEditor opens a temporary file in the user's preferred editor
// and returns the content after the user closes the editor.
func OpenEditor(initialContent string) (string, error) {
	// Get the editor command from environment variables
	editor := getEditorCommand()
	if editor == "" {
		return "", fmt.Errorf("no editor found. Please set EDITOR environment variable")
	}

	// Create a temporary file
	tmpFile, err := ioutil.TempFile("", "devgraph-prompt-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Write initial content if provided
	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}

	// Close the file so the editor can open it
	tmpFile.Close()

	// Prepare the editor command
	var cmd *exec.Cmd
	editorParts := strings.Fields(editor)
	if len(editorParts) == 1 {
		cmd = exec.Command(editorParts[0], tmpFile.Name())
	} else {
		args := append(editorParts[1:], tmpFile.Name())
		cmd = exec.Command(editorParts[0], args...)
	}

	// Set up the command to use the current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the content back
	content, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// getEditorCommand returns the preferred editor command
func getEditorCommand() string {
	// Check common editor environment variables in order of preference
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if editor := os.Getenv(env); editor != "" {
			return editor
		}
	}

	// Fall back to common editors that are likely to be available
	for _, editor := range []string{"nano", "vim", "vi"} {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}
