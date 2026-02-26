// Copyright (C) 2026 Gregory Anders <greg@gpanders.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/emersion/go-scfg"
)

type Config struct {
	HistoryFile   string       `scfg:"history-file"`
	JQCommand     string       `scfg:"jq-bin"`
	HideInputPane bool         `scfg:"hide-input-pane"`
	LibraryPaths  LibraryPaths `scfg:"library-paths"`
	Keymap        Keymap       `scfg:"keymaps"`
}

func DefaultConfig() Config {
	var dataDir string
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		dataDir = dataHome
	} else if home, err := os.UserHomeDir(); err == nil {
		dataDir = filepath.Join(home, ".local", "share")
	}

	var historyFile string
	if dataDir != "" {
		historyFile = filepath.Join(dataDir, "ijq", "history")
	}

	return Config{
		HistoryFile:   historyFile,
		JQCommand:     "jq",
		HideInputPane: false,
		Keymap:        DefaultKeymap(),
	}
}

func NewConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}

		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if err := scfg.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func DefaultConfigPath() (string, error) {
	var config string
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		config = configHome
	} else if home, err := os.UserHomeDir(); err == nil {
		config = filepath.Join(home, ".config")
	} else {
		return "", errors.New("couldn't find path to config file")
	}

	return filepath.Join(config, "ijq", "config"), nil
}
