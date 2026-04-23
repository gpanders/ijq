package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardWrite(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	clipboard := NewClipboard(&buffer)
	n, err := clipboard.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, len("hello"), n)
	assert.Equal(t, "\x1b]52;c;aGVsbG8=\x1b\\", buffer.String())
}

func TestClipboardWriteWithoutTTY(t *testing.T) {
	t.Parallel()

	n, err := NewClipboard(nil).Write([]byte("hello"))
	assert.Zero(t, n)
	assert.ErrorIs(t, err, errNoTTY)
}

func TestClipboardWriteEmpty(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	n, err := NewClipboard(&buffer).Write(nil)
	require.NoError(t, err)
	assert.Zero(t, n)
	assert.Empty(t, buffer.String())
}

func TestClipboardWriteTooLarge(t *testing.T) {
	t.Parallel()

	n, err := NewClipboard(&bytes.Buffer{}).Write(make([]byte, maxClipboardBytes+1))
	assert.Zero(t, n)
	assert.Error(t, err)
}
