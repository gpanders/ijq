package main

import (
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/gpanders/ijq/internal/options"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/ijq-test-config")

	path, err := DefaultConfigPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join("/tmp/ijq-test-config", "ijq", "config"), path)
}

func TestDefaultConfigPathFallsBackToHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmp)

	path, err := DefaultConfigPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, ".config", "ijq", "config"), path)
}

func TestLoadConfigMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	path, err := DefaultConfigPath()
	assert.NoError(t, err)

	cfg, err := NewConfig(path)
	assert.NoError(t, err)

	// Ensure that config has default values when no config file exists
	assert.Equal(t, filepath.Join(tmp, "ijq", "history"), string(cfg.HistoryFile))
	assert.Equal(t, "jq", string(cfg.JQCommand))
	assert.False(t, bool(cfg.HideInputPane))
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")

	raw := `history-file ""
jq-bin /usr/local/bin/jq
hide-input-pane true
library-paths /tmp/modules /opt/jq/modules
keymaps {
	toggle-input-pane Ctrl-T
	save-filter-history Alt+h
	toggle-menu Alt+o
	textview-end g
}
`

	err := os.WriteFile(path, []byte(raw), 0o644)
	assert.NoError(t, err)

	cfg, err := NewConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "", string(cfg.HistoryFile))
	assert.Equal(t, "/usr/local/bin/jq", string(cfg.JQCommand))
	assert.True(t, bool(cfg.HideInputPane))
	assert.Equal(t, options.LibraryPaths{"/tmp/modules", "/opt/jq/modules"}, cfg.LibraryPaths)

	assert.Equal(t, KeyBindings{{key: tcell.KeyCtrlT}}, cfg.Keymap.ToggleInputPane)
	assert.Equal(t, KeyBindings{{key: tcell.KeyRune, rune: 'h', mods: tcell.ModAlt}}, cfg.Keymap.SaveFilterHistory)
	assert.Equal(t, KeyBindings{{key: tcell.KeyRune, rune: 'o', mods: tcell.ModAlt}}, cfg.Keymap.ToggleMenu)
	assert.Equal(t, KeyBindings{{key: tcell.KeyRune, rune: 'g'}}, cfg.Keymap.TextviewEnd)
	assert.Equal(t, DefaultKeymap().SubmitFilter, cfg.Keymap.SubmitFilter)
}

func TestLoadConfigInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")

	err := os.WriteFile(path, []byte("keymaps {"), 0o644)
	assert.NoError(t, err)

	_, err = NewConfig(path)
	assert.Error(t, err)
}
