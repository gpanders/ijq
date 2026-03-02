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
	"os/exec"
	"strings"
	"testing"

	"codeberg.org/gpanders/ijq/internal/options"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
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
	assert.Contains(t, help, "Alt+m")
	assert.Contains(t, help, "Ctrl-C")
	assert.Contains(t, help, "Ctrl+s")
}
