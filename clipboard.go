// Copyright (C) 2026 Gregory Anders <greg@gpanders.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"sync"
)

type Clipboard struct {
	mutex sync.RWMutex
	tty   io.Writer
}

var _ io.Writer = &Clipboard{}

func (c *Clipboard) Write(buf []byte) (int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.tty == nil {
		return 0, nil
	}

	out := fmt.Appendf(nil, "\x1b]52;c;%s\x1b\\", base64.StdEncoding.EncodeToString(buf))
	if _, err := c.tty.Write(out); err != nil {
		return 0, fmt.Errorf("failed to write OSC 52 escape sequence: %w", err)
	}

	return len(buf), nil
}

func (c *Clipboard) SetTTY(tty io.Writer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.tty = tty
}
