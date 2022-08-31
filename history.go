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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type history struct {
	path  string
	lines []string
	Items []string
}

var colorTagRe = regexp.MustCompile(`\[([^]]+)\]`)

// Escape bracketed expressions (like [this]) because they are interpreted as
// color tags by tview.
// See https://godocs.io/github.com/rivo/tview#hdr-Colors
func escape(s string) string {
	return colorTagRe.ReplaceAllString(s, "[$1[]")
}

func (h *history) Init(path string) error {
	h.path = path

	filebytes, err := ioutil.ReadFile(path)
	if err != nil {
		// If the history file doesn't exist, then
		// return an empty history.
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else {
			return fmt.Errorf("error retrieving history: %w", err)
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(filebytes))
	for scanner.Scan() {
		line := scanner.Text()
		h.lines = append(h.lines, line)
		h.Items = append(h.Items, escape(line))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf(
			"error retrieving history: %w", err,
		)
	}

	return nil
}

func (h *history) Add(expression string) error {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}

	if h.path == "" {
		return nil
	}

	// Don't continue with adding the expression if it is saved in history
	// already.
	if contains(h.lines, expression) {
		return nil
	}

	h.lines = append(h.lines, expression)
	h.Items = append(h.Items, escape(expression))

	file, err := h.openFile()
	if err != nil {
		return fmt.Errorf("error opening history for writing: %w", err)
	}

	fmt.Fprintln(file, expression)

	if err = file.Close(); err != nil {
		return fmt.Errorf("error closing history file: %w", err)
	}

	return nil
}

func (h *history) openFile() (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(h.path), os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(h.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func contains(arr []string, elem string) bool {
	for _, v := range arr {
		if elem == v {
			return true
		}
	}

	return false
}
