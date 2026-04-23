package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardWrite(t *testing.T) {
	var buffer bytes.Buffer

	clipboard := &Clipboard{}
	clipboard.SetTTY(&buffer)
	n, err := clipboard.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, len("hello"), n)
	assert.Equal(t, "\x1b]52;c;aGVsbG8=\x1b\\", buffer.String())
}
