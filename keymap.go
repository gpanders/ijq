// Copyright (C) 2026 Gregory Anders <greg@gpanders.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type KeyBinding struct {
	key  tcell.Key
	rune rune
	mods tcell.ModMask
}

func (k *KeyBinding) UnmarshalText(text []byte) error {
	trimmed := strings.TrimSpace(string(text))
	if trimmed == "" {
		return errors.New("key binding cannot be empty")
	}

	modsPart, base := splitKeyBinding(trimmed)
	if base == "" {
		return errors.New("key binding modifier is missing a key")
	}

	mods := tcell.ModNone
	if modsPart != "" {
		for _, modifier := range strings.Split(modsPart, "+") {
			modifier = strings.ToLower(strings.TrimSpace(modifier))
			switch modifier {
			case "shift":
				mods |= tcell.ModShift
			case "alt":
				mods |= tcell.ModAlt
			case "ctrl", "control":
				mods |= tcell.ModCtrl
			case "meta":
				mods |= tcell.ModMeta
			default:
				return fmt.Errorf("unknown modifier %q", modifier)
			}
		}
	}

	base = strings.TrimSpace(base)
	if key, ok := keyNameLookup[strings.ToLower(base)]; ok {
		k.key = key
		k.mods = mods
		return nil
	}

	if mods&tcell.ModCtrl != 0 {
		ctrlName := "ctrl-" + strings.ToLower(base)
		if key, ok := keyNameLookup[ctrlName]; ok {
			k.key = key
			k.mods = mods &^ tcell.ModCtrl
			return nil
		}
	}

	runes := []rune(base)
	if len(runes) == 1 {
		r := runes[0]
		if mods&tcell.ModShift != 0 {
			upper := unicode.ToUpper(r)
			if upper != r {
				r = upper
				mods &^= tcell.ModShift
			}
		}

		k.key = tcell.KeyRune
		k.rune = r
		k.mods = mods
		return nil
	}

	return fmt.Errorf("unknown key %q", base)
}

func (binding KeyBinding) Matches(event *tcell.EventKey) bool {
	if event.Key() != binding.key {
		return false
	}

	if binding.key == tcell.KeyRune && event.Rune() != binding.rune {
		return false
	}

	return event.Modifiers() == binding.mods
}

type KeyBindings []KeyBinding

func (bindings KeyBindings) Matches(event *tcell.EventKey) bool {
	for _, binding := range bindings {
		if binding.Matches(event) {
			return true
		}
	}

	return false
}

type Keymap struct {
	SubmitFilter KeyBindings `scfg:"submit-filter"`
	MoveDown     KeyBindings `scfg:"move-down"`
	MoveUp       KeyBindings `scfg:"move-up"`
	PageDown     KeyBindings `scfg:"page-down"`
	LineStart    KeyBindings `scfg:"line-start"`
	LineEnd      KeyBindings `scfg:"line-end"`
	HalfPageUp   KeyBindings `scfg:"half-page-up"`
	HalfPageDown KeyBindings `scfg:"half-page-down"`

	FilterCursorRight KeyBindings `scfg:"filter-cursor-right"`
	FilterCursorLeft  KeyBindings `scfg:"filter-cursor-left"`

	FocusInputPaneUp   KeyBindings `scfg:"focus-input-pane-up"`
	FocusInputPaneLeft KeyBindings `scfg:"focus-input-pane-left"`
	FocusOutputPane    KeyBindings `scfg:"focus-output-pane"`
	FocusFilterInput   KeyBindings `scfg:"focus-filter-input"`
	NextFocus          KeyBindings `scfg:"next-focus"`
	PreviousFocus      KeyBindings `scfg:"previous-focus"`
	ToggleInputPane    KeyBindings `scfg:"toggle-input-pane"`

	TextviewPageUp    KeyBindings `scfg:"textview-page-up"`
	TextviewPageUpAlt KeyBindings `scfg:"textview-page-up-alt"`
	TextviewPageDown  KeyBindings `scfg:"textview-page-down"`
	TextviewEnd       KeyBindings `scfg:"textview-end"`
}

func DefaultKeymap() Keymap {
	return Keymap{
		SubmitFilter: KeyBindings{{key: tcell.KeyEnter}},
		MoveDown:     KeyBindings{{key: tcell.KeyCtrlN}},
		MoveUp:       KeyBindings{{key: tcell.KeyCtrlP}},
		PageDown:     KeyBindings{{key: tcell.KeyCtrlV}},
		LineStart:    KeyBindings{{key: tcell.KeyCtrlA}, {key: tcell.KeyRune, rune: '0'}},
		LineEnd:      KeyBindings{{key: tcell.KeyCtrlE}, {key: tcell.KeyRune, rune: '$'}},
		HalfPageUp:   KeyBindings{{key: tcell.KeyCtrlU}, {key: tcell.KeyRune, rune: 'u'}},
		HalfPageDown: KeyBindings{{key: tcell.KeyCtrlD}, {key: tcell.KeyRune, rune: 'd'}},

		FilterCursorRight: KeyBindings{{key: tcell.KeyCtrlF}},
		FilterCursorLeft:  KeyBindings{{key: tcell.KeyCtrlB}},

		FocusInputPaneUp:   KeyBindings{{key: tcell.KeyUp, mods: tcell.ModShift}},
		FocusInputPaneLeft: KeyBindings{{key: tcell.KeyLeft, mods: tcell.ModShift}},
		FocusOutputPane:    KeyBindings{{key: tcell.KeyRight, mods: tcell.ModShift}},
		FocusFilterInput:   KeyBindings{{key: tcell.KeyDown, mods: tcell.ModShift}},
		NextFocus:          KeyBindings{{key: tcell.KeyTab}},
		PreviousFocus:      KeyBindings{{key: tcell.KeyBacktab}},
		ToggleInputPane:    KeyBindings{{key: tcell.KeyCtrlO}},

		TextviewPageUp:    KeyBindings{{key: tcell.KeyRune, rune: 'b'}},
		TextviewPageUpAlt: KeyBindings{{key: tcell.KeyRune, rune: 'v', mods: tcell.ModAlt}},
		TextviewPageDown:  KeyBindings{{key: tcell.KeyRune, rune: 'f'}},
		TextviewEnd:       KeyBindings{{key: tcell.KeyRune, rune: 'G'}},
	}
}

func splitKeyBinding(value string) (mods string, base string) {
	if value == "+" {
		return "", "+"
	}

	if strings.HasSuffix(value, "++") {
		return strings.TrimSpace(value[:len(value)-2]), "+"
	}

	if idx := strings.LastIndex(value, "+"); idx != -1 {
		return strings.TrimSpace(value[:idx]), strings.TrimSpace(value[idx+1:])
	}

	base = value
	if _, ok := keyNameLookup[strings.ToLower(base)]; ok {
		return "", base
	}

	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return "", base
	}

	modifiers := make([]string, 0, len(parts)-1)
	for _, part := range parts[:len(parts)-1] {
		modifier := strings.TrimSpace(part)
		if !isModifier(modifier) {
			return "", base
		}

		modifiers = append(modifiers, modifier)
	}

	return strings.Join(modifiers, "+"), strings.TrimSpace(parts[len(parts)-1])
}

func isModifier(value string) bool {
	switch strings.ToLower(value) {
	case "shift", "alt", "ctrl", "control", "meta":
		return true
	default:
		return false
	}
}

var keyNameLookup = func() map[string]tcell.Key {
	lookup := make(map[string]tcell.Key, len(tcell.KeyNames)+5)
	for key, name := range tcell.KeyNames {
		lookup[strings.ToLower(name)] = key
	}

	lookup["return"] = tcell.KeyEnter
	lookup["pageup"] = tcell.KeyPgUp
	lookup["pagedown"] = tcell.KeyPgDn
	lookup["pgdown"] = tcell.KeyPgDn
	lookup["shift-tab"] = tcell.KeyBacktab

	return lookup
}()
