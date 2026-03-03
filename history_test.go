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
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randomFilename(namebase string) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	if namebase == "" {
		namebase = "file"
	}

	return namebase + "." + strconv.Itoa(random.Int())
}

func makeHistoryFilename() string {
	return randomFilename("./history")
}

func TestHistoryAddNoFilename(t *testing.T) {
	var h history
	err := h.Add("foo")
	assert.NoError(t, err)
	assert.Nil(t, err)
	assert.Empty(t, h.Items)
}

func TestHistoryAddEmptyString(t *testing.T) {
	histfile := makeHistoryFilename()
	var h history
	h.Init(histfile)
	err := h.Add("")
	assert.NoError(t, err)
	assert.Nil(t, err)
	assert.NoFileExists(t, histfile)
}

func TestHistoryAdd(t *testing.T) {
	histFile := makeHistoryFilename()

	before := "one\ntwo\n"
	after := "one\ntwo\nthree\n"

	err := os.WriteFile(histFile, []byte(before), 0644)
	assert.NoError(t, err)

	var h history
	h.Init(histFile)

	err = h.Add("three")
	assert.NoError(t, err)

	contents, err := os.ReadFile(histFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte(after), contents)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistoryAddRepeating(t *testing.T) {
	histFile := makeHistoryFilename()

	contents := "one\ntwo\n"

	err := os.WriteFile(histFile, []byte(contents), 0644)
	assert.NoError(t, err)

	var h history
	h.Init(histFile)

	err = h.Add("one")
	assert.NoError(t, err)

	retrieved, err := os.ReadFile(histFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte(contents), retrieved)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistoryWithinSubDir(t *testing.T) {
	rootDir := "./testdata/myroot"

	histFile := path.Join(rootDir, "myhistory")

	var h history
	h.Init(histFile)

	err := h.Add("one")
	assert.NoError(t, err)
	assert.FileExists(t, histFile)

	assert.NoError(t, os.RemoveAll(rootDir))
}

func TestHistoryGetMissingFile(t *testing.T) {
	historyFile := "./this.does.not.exist"

	var h history
	h.Init(historyFile)
	assert.NoFileExists(t, historyFile)
}

func TestHistory(t *testing.T) {
	histFile := makeHistoryFilename()

	var h history
	h.Init(histFile)

	var err error

	err = h.Add("one")
	assert.NoError(t, err)
	err = h.Add("two")
	assert.NoError(t, err)
	err = h.Add("three")
	assert.NoError(t, err)

	assert.NoError(t, err)

	assert.Equal(
		t,
		h.Items,
		[]string{"one", "two", "three"},
	)

	// Add new after Get
	err = h.Add("four")
	assert.NoError(t, err)

	assert.Equal(
		t,
		h.Items,
		[]string{"one", "two", "three", "four"},
	)

	// Attempt to add item already in history
	err = h.Add("one")
	assert.NoError(t, err)

	assert.Equal(
		t,
		h.Items,
		[]string{"one", "two", "three", "four"},
	)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistoryAddIfMissingStatus(t *testing.T) {
	histFile := makeHistoryFilename()
	defer os.Remove(histFile)

	var h history
	err := h.Init(histFile)
	assert.NoError(t, err)

	added, err := h.AddIfMissing("foo")
	assert.NoError(t, err)
	assert.True(t, added)

	added, err = h.AddIfMissing("foo")
	assert.NoError(t, err)
	assert.False(t, added)
}

func TestHistoryDeleteAt(t *testing.T) {
	histFile := makeHistoryFilename()
	defer os.Remove(histFile)

	err := os.WriteFile(histFile, []byte("one\ntwo\nthree\n"), 0644)
	assert.NoError(t, err)

	var h history
	err = h.Init(histFile)
	assert.NoError(t, err)

	err = h.DeleteAt(1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"one", "three"}, h.Items)

	contents, err := os.ReadFile(histFile)
	assert.NoError(t, err)
	assert.Equal(t, "one\nthree\n", string(contents))
}

func TestHistoryDeleteAtInvalidIndex(t *testing.T) {
	var h history
	h.Items = []string{"one", "two"}

	assert.Error(t, h.DeleteAt(-1))
	assert.Error(t, h.DeleteAt(2))
}

func TestHistoryDeleteAtRewriteSetsFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not preserve unix permission bits")
	}

	histFile := path.Join(t.TempDir(), "history")
	err := os.WriteFile(histFile, []byte("one\ntwo\n"), 0o600)
	assert.NoError(t, err)

	var h history
	err = h.Init(histFile)
	assert.NoError(t, err)

	err = h.DeleteAt(0)
	assert.NoError(t, err)

	info, err := os.Stat(histFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestHistoryEntriesReturnsCopy(t *testing.T) {
	h := history{Items: []string{"one", "two"}}

	entries := h.Entries()
	entries[0] = "changed"

	assert.Equal(t, []string{"one", "two"}, h.Items)
}
