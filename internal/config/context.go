package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const contextFileName = "context"

// ContextManager manages the active conference context stored on disk.
type ContextManager struct {
	contextDir string
}

// NewContextManager creates a new ContextManager that stores context in the given directory.
func NewContextManager(contextDir string) *ContextManager {
	return &ContextManager{contextDir: contextDir}
}

// GetActiveConference returns the currently active conference slug.
// Returns empty string and nil error if no context is set.
func (cm *ContextManager) GetActiveConference() (string, error) {
	data, err := os.ReadFile(cm.GetContextFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read active conference: %w", err)
	}

	slug := strings.TrimSpace(string(data))
	return slug, nil
}

// SetActiveConference saves the given slug as the active conference context.
func (cm *ContextManager) SetActiveConference(slug string) error {
	if err := os.MkdirAll(cm.contextDir, 0o700); err != nil {
		return fmt.Errorf("failed to save active conference: %w", err)
	}

	if err := os.WriteFile(cm.GetContextFilePath(), []byte(slug+"\n"), 0o600); err != nil {
		return fmt.Errorf("failed to save active conference: %w", err)
	}

	return nil
}

// ClearActiveConference removes the active conference context.
func (cm *ContextManager) ClearActiveConference() error {
	err := os.Remove(cm.GetContextFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to clear active conference: %w", err)
	}

	return nil
}

// GetContextFilePath returns the full path to the context file.
func (cm *ContextManager) GetContextFilePath() string {
	return filepath.Join(cm.contextDir, contextFileName)
}
