// Copyright (C) 2021 Gregory Anders <greg@gpanders.com>
// Copyright (C) 2021 Herby Gillot <herby.gillot@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/gpanders/ijq/internal/options"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentReadFrom(t *testing.T) {
	testMsg := "hello world"
	testReader := strings.NewReader(testMsg)

	doc := &Document{}

	readCount, err := doc.ReadFrom(testReader)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), int(readCount))
}

func TestDocumentWriteTo(t *testing.T) {
	testMsg := "hello world"
	testReader := strings.NewReader(testMsg)

	doc := &Document{
		filter:  "-",
		options: options.Options{JQCommand: "cat"},
		ctx:     context.Background(),
	}

	readCount, err := doc.ReadFrom(testReader)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), int(readCount))

	buffer := bytes.Buffer{}

	writeCount, err := doc.WriteTo(&buffer)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(writeCount))

	assert.Equal(t, testMsg, buffer.String())
}

func TestDocumentWithFilterPreservesFields(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	opts := options.Options{
		CompactOutput: true,
		JQCommand:     "jq",
	}
	original := Document{
		input:   `{"foo":1}`,
		filter:  ".",
		options: opts,
		config:  cfg,
		ctx:     ctx,
	}

	next := original.WithFilter(".foo")

	assert.Equal(t, original.input, next.input)
	assert.Equal(t, ".foo", next.filter)
	assert.Equal(t, original.options, next.options)
	assert.Equal(t, original.config, next.config)
	assert.NotNil(t, next.ctx)
}

func TestDocumentExecError(t *testing.T) {
	testMsg := "hello world"
	testReader := strings.NewReader(testMsg)

	doc := &Document{
		options: options.Options{JQCommand: "./testdata/caterror"},
		ctx:     context.Background(),
	}

	readCount, err := doc.ReadFrom(testReader)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), int(readCount))

	buffer := bytes.Buffer{}

	writeCount, err := doc.WriteTo(&buffer)
	assert.Error(t, err)
	assert.Equal(t, 0, int(writeCount))

	exiterr, ok := err.(*exec.ExitError)
	assert.True(t, ok)
	assert.NotNil(t, exiterr)
	assert.Equal(t, testMsg, string(exiterr.Stderr))

	assert.Empty(t, buffer.String())
}

func TestNormalizeOverlayEvent(t *testing.T) {
	keymap := DefaultKeymap()

	event := normalizeOverlayEvent(tcell.NewEventKey(tcell.KeyCtrlN, ' ', tcell.ModNone), keymap)
	assert.Equal(t, tcell.KeyDown, event.Key())

	event = normalizeOverlayEvent(tcell.NewEventKey(tcell.KeyCtrlP, ' ', tcell.ModNone), keymap)
	assert.Equal(t, tcell.KeyUp, event.Key())

	event = normalizeOverlayEvent(tcell.NewEventKey(tcell.KeyCtrlV, ' ', tcell.ModNone), keymap)
	assert.Equal(t, tcell.KeyPgDn, event.Key())

	event = normalizeOverlayEvent(tcell.NewEventKey(tcell.KeyCtrlD, ' ', tcell.ModNone), keymap)
	assert.Equal(t, tcell.KeyPgDn, event.Key())

	event = normalizeOverlayEvent(tcell.NewEventKey(tcell.KeyCtrlU, ' ', tcell.ModNone), keymap)
	assert.Equal(t, tcell.KeyPgUp, event.Key())

	original := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	event = normalizeOverlayEvent(original, keymap)
	assert.Equal(t, original, event)
}

func TestBuildMainHelpTextUsesConfiguredBindings(t *testing.T) {
	keymap := DefaultKeymap()
	keymap.ToggleMenu = KeyBindings{{key: tcell.KeyRune, rune: 'm', mods: tcell.ModAlt}}
	keymap.SubmitFilter = KeyBindings{{key: tcell.KeyRune, rune: 's', mods: tcell.ModCtrl}}

	help := buildMainHelpText(keymap)
	assert.Contains(t, help, "Alt-m")
	assert.Contains(t, help, "Ctrl-C")
	assert.Contains(t, help, "Ctrl-s")
}

func TestParseArgsLoadsFilterFromFile(t *testing.T) {
	filterFile := filepath.Join(t.TempDir(), "filter.jq")
	require.NoError(t, os.WriteFile(filterFile, []byte(".foo\n"), 0o644))

	oldArgs := os.Args
	os.Args = []string{"ijq", "-f", filterFile, "input.json"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	opts := options.Options{}
	filter, args := parseArgs(&opts)

	assert.Equal(t, ".foo\n", filter)
	assert.Equal(t, []string{"input.json"}, args)
}

func TestParseArgsTreatsFirstArgAsFilterWithNullInput(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"ijq", ".items[]"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	opts := options.Options{NullInput: true}
	filter, args := parseArgs(&opts)

	assert.Equal(t, ".items[]", filter)
	assert.Empty(t, args)
}

func TestParseArgsTreatsFirstArgAsFilterWhenMultiplePositionals(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"ijq", ".foo", "file1.json", "file2.json"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	opts := options.Options{}
	filter, args := parseArgs(&opts)

	assert.Equal(t, ".foo", filter)
	assert.Equal(t, []string{"file1.json", "file2.json"}, args)
}

func TestParseArgsVersionFlagPrintsAndExits(t *testing.T) {
	if os.Getenv("IJQ_PARSEARGS_VERSION_HELPER") == "1" {
		oldArgs := os.Args
		oldVersion := Version
		defer func() {
			os.Args = oldArgs
			Version = oldVersion
		}()

		Version = "test-version"
		os.Args = []string{"ijq", "-V"}
		opts := options.Options{}
		parseArgs(&opts)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseArgsVersionFlagPrintsAndExits")
	cmd.Env = append(os.Environ(), "IJQ_PARSEARGS_VERSION_HELPER=1")

	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(out), "ijq test-version")
}
