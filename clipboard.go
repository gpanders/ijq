// Copyright (C) 2026 Gregory Anders <greg@gpanders.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
)

const maxClipboardBytes = 1 << 20

var errNoTTY = errors.New("terminal does not expose a tty")

type Clipboard struct {
	mutex sync.Mutex
	tty   io.Writer
}

var _ io.Writer = &Clipboard{}

func NewClipboard(tty io.Writer) *Clipboard {
	return &Clipboard{tty: tty}
}

func (c *Clipboard) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}

	if len(buf) > maxClipboardBytes {
		return 0, fmt.Errorf("clipboard content is %d bytes; limit is %d", len(buf), maxClipboardBytes)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.tty == nil {
		return 0, errNoTTY
	}

	out := fmt.Appendf(nil, "\x1b]52;c;%s\x1b\\", base64.StdEncoding.EncodeToString(buf))
	if _, err := c.tty.Write(out); err != nil {
		return 0, fmt.Errorf("failed to write OSC 52 escape sequence: %w", err)
	}

	return len(buf), nil
}
