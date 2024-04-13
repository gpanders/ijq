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

	"github.com/stretchr/testify/assert"
)

func TestOptionsToSlice(t *testing.T) {
	opt := &Options{}

	opt.compact = true
	assert.Contains(t, opt.ToSlice(), "-c")
	opt.compact = false
	assert.NotContains(t, opt.ToSlice(), "-c")

	opt.nullInput = true
	assert.Contains(t, opt.ToSlice(), "-n")
	opt.nullInput = false
	assert.NotContains(t, opt.ToSlice(), "-n")

	opt.slurp = true
	assert.Contains(t, opt.ToSlice(), "-s")
	opt.slurp = false
	assert.NotContains(t, opt.ToSlice(), "-s")

	opt.rawOutput = true
	assert.Contains(t, opt.ToSlice(), "-r")
	opt.rawOutput = false
	assert.NotContains(t, opt.ToSlice(), "-r")

	opt.rawInput = true
	assert.Contains(t, opt.ToSlice(), "-R")
	opt.rawInput = false
	assert.NotContains(t, opt.ToSlice(), "-R")

	opt.monochrome = true
	assert.Contains(t, opt.ToSlice(), "-M")
	opt.monochrome = false
	assert.NotContains(t, opt.ToSlice(), "-M")

	opt.forceColor = true
	assert.Contains(t, opt.ToSlice(), "-C")
	opt.forceColor = false
	assert.NotContains(t, opt.ToSlice(), "-C")

	opt.sortKeys = true
	assert.Contains(t, opt.ToSlice(), "-S")
	opt.sortKeys = false
	assert.NotContains(t, opt.ToSlice(), "-S")
}

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
		filter: "-",
		options: Options{
			command: "cat",
		},
		ctx: context.Background(),
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
		options: Options{
			command: "./testdata/caterror",
		},
		ctx: context.Background(),
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
