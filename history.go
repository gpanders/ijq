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
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type history struct {
	path  string
	Items []string
}

func (h *history) Init(path string) error {
	h.path = path

	filebytes, err := os.ReadFile(path)
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
		h.Items = append(h.Items, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf(
			"error retrieving history: %w", err,
		)
	}

	return nil
}

func (h *history) Add(expression string) error {
	_, err := h.AddIfMissing(expression)
	return err
}

func (h *history) AddIfMissing(expression string) (bool, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return false, nil
	}

	if h.path == "" {
		return false, nil
	}

	// Don't continue with adding the expression if it is saved in history
	// already.
	if slices.Contains(h.Items, expression) {
		return false, nil
	}

	file, err := h.openFile()
	if err != nil {
		return false, fmt.Errorf("error opening history for writing: %w", err)
	}

	if _, err = fmt.Fprintln(file, expression); err != nil {
		file.Close()
		return false, fmt.Errorf("error writing history file: %w", err)
	}

	if err = file.Close(); err != nil {
		return false, fmt.Errorf("error closing history file: %w", err)
	}

	h.Items = append(h.Items, expression)

	return true, nil
}

func (h *history) DeleteAt(index int) error {
	if index < 0 || index >= len(h.Items) {
		return fmt.Errorf("history index out of range")
	}

	nextItems := append([]string(nil), h.Items[:index]...)
	nextItems = append(nextItems, h.Items[index+1:]...)

	if h.path == "" {
		h.Items = nextItems
		return nil
	}

	if err := h.rewrite(nextItems); err != nil {
		return fmt.Errorf("error rewriting history: %w", err)
	}

	h.Items = nextItems

	return nil
}

func (h *history) openFile() (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(h.path), os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(h.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (h *history) Entries() []string {
	return append([]string(nil), h.Items...)
}

func (h *history) rewrite(items []string) (rerr error) {
	if err := os.MkdirAll(filepath.Dir(h.path), os.ModePerm); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(h.path), ".history-*")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	defer func() {
		if rerr != nil {
			tmpFile.Close()
			os.Remove(tmpName)
		}
	}()

	for _, item := range items {
		if _, err := io.WriteString(tmpFile, item+"\n"); err != nil {
			return err
		}
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, h.path); err != nil {
		return err
	}

	if err := os.Chmod(h.path, 0o644); err != nil {
		return err
	}

	return nil
}
