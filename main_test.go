package main

import (
	"bytes"
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
	}

	readCount, err := doc.ReadFrom(testReader)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), int(readCount))

	buffer := bytes.Buffer{}

	writeCount, err := doc.WriteTo(&buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testMsg), int(writeCount))

	assert.Equal(t, testMsg, buffer.String())
}

func TestDocumentExecError(t *testing.T) {
	testMsg := "hello world"
	testReader := strings.NewReader(testMsg)

	doc := &Document{
		options: Options{
			command: "./testdata/caterror",
		},
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
