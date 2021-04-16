package main

import (
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
