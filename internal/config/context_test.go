package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextManager_SetAndGet(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	err := cm.SetActiveConference("socrates-26")
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Equal(t, "socrates-26", slug)
}

func TestContextManager_GetNoFile(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Empty(t, slug)
}

func TestContextManager_GetEmptyFile(t *testing.T) {
	dir := t.TempDir()
	cm := NewContextManager(dir)

	err := os.WriteFile(filepath.Join(dir, contextFileName), []byte(""), 0o600)
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Empty(t, slug)
}

func TestContextManager_SetOverwrites(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	err := cm.SetActiveConference("conf-a")
	require.NoError(t, err)

	err = cm.SetActiveConference("conf-b")
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Equal(t, "conf-b", slug)
}

func TestContextManager_SetCreatesDirectory(t *testing.T) {
	baseDir := t.TempDir()
	contextDir := filepath.Join(baseDir, "nested", "dir")
	cm := NewContextManager(contextDir)

	err := cm.SetActiveConference("socrates-26")
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Equal(t, "socrates-26", slug)
}

func TestContextManager_ClearSuccess(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	err := cm.SetActiveConference("socrates-26")
	require.NoError(t, err)

	err = cm.ClearActiveConference()
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Empty(t, slug)
}

func TestContextManager_ClearNoFile(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	err := cm.ClearActiveConference()
	require.NoError(t, err)
}

func TestContextManager_GetContextFilePath(t *testing.T) {
	dir := "/some/test/dir"
	cm := NewContextManager(dir)

	assert.Equal(t, filepath.Join(dir, "context"), cm.GetContextFilePath())
}

func TestContextManager_SetActiveConference_InvalidSlugs(t *testing.T) {
	tests := []struct {
		name string
		slug string
	}{
		{name: "empty", slug: ""},
		{name: "uppercase", slug: "SOCRATES-26"},
		{name: "path separator", slug: "../evil"},
		{name: "spaces", slug: "my conference"},
		{name: "starts with hyphen", slug: "-invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewContextManager(t.TempDir())
			err := cm.SetActiveConference(tt.slug)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid conference slug")
		})
	}
}

func TestContextManager_SetActiveConference_ValidSlug(t *testing.T) {
	cm := NewContextManager(t.TempDir())

	err := cm.SetActiveConference("socrates-26")
	require.NoError(t, err)

	slug, err := cm.GetActiveConference()
	require.NoError(t, err)
	assert.Equal(t, "socrates-26", slug)
}
