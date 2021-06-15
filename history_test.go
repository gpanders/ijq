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
	"io/ioutil"
	"math/rand"
	"os"
	"path"
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
	h := &history{}

	err := h.Add("foo")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyPath)
}

func TestHistoryAddEmptyString(t *testing.T) {
	histfile := makeHistoryFilename()
	h := &history{Path: histfile}
	err := h.Add("")
	assert.NoError(t, err)
	assert.Nil(t, err)
	assert.NoFileExists(t, histfile)
}

func TestContains(t *testing.T) {
	things := []string{"one", "two", "three"}

	assert.True(t, contains(things, "one"))
	assert.True(t, contains(things, "two"))
	assert.True(t, contains(things, "three"))
	assert.False(t, contains(things, "four"))
}

func TestHistoryAdd(t *testing.T) {
	histFile := makeHistoryFilename()

	before := "one\ntwo\n"
	after := "one\ntwo\nthree\n"

	err := ioutil.WriteFile(histFile, []byte(before), 0644)
	assert.NoError(t, err)

	h := &history{Path: histFile}

	err = h.Add("three")
	assert.NoError(t, err)

	contents, err := ioutil.ReadFile(histFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte(after), contents)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistoryAddRepeating(t *testing.T) {
	histFile := makeHistoryFilename()

	contents := "one\ntwo\n"

	err := ioutil.WriteFile(histFile, []byte(contents), 0644)
	assert.NoError(t, err)

	h := &history{Path: histFile}

	err = h.Add("one")
	assert.NoError(t, err)

	retrieved, err := ioutil.ReadFile(histFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte(contents), retrieved)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistoryWithinSubDir(t *testing.T) {
	rootDir := "./testdata/myroot"

	histFile := path.Join(rootDir, "myhistory")

	h := &history{Path: histFile}

	err := h.Add("one")
	assert.NoError(t, err)
	assert.FileExists(t, histFile)

	assert.NoError(t, os.RemoveAll(rootDir))
}

func TestHistoryGetMissingFile(t *testing.T) {
	historyFile := "./this.does.not.exist"

	h := &history{Path: historyFile}

	saved, err := h.Get()
	assert.NoError(t, err)
	assert.Nil(t, saved)
	assert.NoFileExists(t, historyFile)
}

func TestHistoryGet(t *testing.T) {
	histFile := makeHistoryFilename()

	fileContents := "one\ntwo\nthree\n"

	expressions := []string{"one", "two", "three"}

	err := ioutil.WriteFile(histFile, []byte(fileContents), 0644)
	assert.NoError(t, err)

	h := &history{Path: histFile}

	retrieved, err := h.Get()
	assert.NoError(t, err)
	assert.Equal(t, expressions, retrieved)

	assert.NoError(t, os.Remove(histFile))
}

func TestHistory(t *testing.T) {
	histFile := makeHistoryFilename()

	h := &history{Path: histFile}

	var err error

	err = h.Add("one")
	assert.NoError(t, err)
	err = h.Add("two")
	assert.NoError(t, err)
	err = h.Add("three")
	assert.NoError(t, err)

	retrieved, err := h.Get()
	assert.NoError(t, err)

	assert.Equal(
		t,
		retrieved,
		[]string{"one", "two", "three"},
	)

	// Add new after Get
	err = h.Add("four")
	assert.NoError(t, err)

	retrieved, err = h.Get()
	assert.NoError(t, err)

	assert.Equal(
		t,
		retrieved,
		[]string{"one", "two", "three", "four"},
	)

	// Attempt to add item already in history
	err = h.Add("one")
	assert.NoError(t, err)

	retrieved, err = h.Get()
	assert.NoError(t, err)

	assert.Equal(
		t,
		retrieved,
		[]string{"one", "two", "three", "four"},
	)

	assert.NoError(t, os.Remove(histFile))
}
