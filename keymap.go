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
		for modifier := range strings.SplitSeq(modsPart, "+") {
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

	if event.Modifiers() == binding.mods {
		return true
	}

	if binding.mods == tcell.ModNone && isControlKey(binding.key) {
		mods := event.Modifiers()
		return mods == tcell.ModCtrl || mods == tcell.ModCtrl|tcell.ModShift
	}

	return false
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

func (binding KeyBinding) String() string {
	if binding.key == tcell.KeyCtrlUnderscore {
		return "Ctrl-/"
	}

	if binding.key == tcell.KeyRune {
		base := string(binding.rune)
		if binding.rune == ' ' {
			base = "Space"
		}

		return formatKeyLabel(binding.mods, base)
	}

	base := keyName(binding.key)
	if strings.HasPrefix(base, "Ctrl-") || strings.HasPrefix(base, "Alt+") || strings.HasPrefix(base, "Shift+") {
		return base
	}

	return formatKeyLabel(binding.mods, base)
}

func (bindings KeyBindings) PrimaryString() string {
	if len(bindings) == 0 {
		return ""
	}

	return bindings[0].String()
}

func formatKeyLabel(mods tcell.ModMask, base string) string {
	if mods == tcell.ModNone {
		return base
	}

	parts := make([]string, 0, 5)
	if mods&tcell.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if mods&tcell.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mods&tcell.ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if mods&tcell.ModMeta != 0 {
		parts = append(parts, "Meta")
	}

	parts = append(parts, base)
	return strings.Join(parts, "+")
}

func keyName(key tcell.Key) string {
	switch key {
	case tcell.KeyEnter:
		return "Enter"
	case tcell.KeyEsc:
		return "Esc"
	case tcell.KeyTab:
		return "Tab"
	case tcell.KeyBacktab:
		return "Shift-Tab"
	case tcell.KeyPgUp:
		return "PageUp"
	case tcell.KeyPgDn:
		return "PageDown"
	}

	if name, ok := tcell.KeyNames[key]; ok {
		return name
	}

	return fmt.Sprintf("Key(%d)", key)
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
	ToggleMenu         KeyBindings `scfg:"toggle-menu"`

	TextviewPageUp   KeyBindings `scfg:"textview-page-up"`
	TextviewPageDown KeyBindings `scfg:"textview-page-down"`
	TextviewEnd      KeyBindings `scfg:"textview-end"`
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
		ToggleMenu: KeyBindings{
			{key: tcell.KeyCtrlUnderscore},
			{key: tcell.KeyRune, rune: '/', mods: tcell.ModCtrl},
			{key: tcell.KeyRune, rune: '?', mods: tcell.ModCtrl},
			{key: tcell.KeyRune, rune: '_', mods: tcell.ModCtrl},
		},

		TextviewPageUp:   KeyBindings{{key: tcell.KeyRune, rune: 'b'}, {key: tcell.KeyRune, rune: 'v', mods: tcell.ModAlt}},
		TextviewPageDown: KeyBindings{{key: tcell.KeyRune, rune: 'f'}},
		TextviewEnd:      KeyBindings{{key: tcell.KeyRune, rune: 'G'}},
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

func isControlKey(key tcell.Key) bool {
	if key < tcell.KeyCtrlSpace || key > tcell.KeyCtrlUnderscore {
		return false
	}

	switch key {
	case tcell.KeyBackspace, tcell.KeyTab, tcell.KeyEsc, tcell.KeyEnter:
		return false
	default:
		return true
	}
}

var keyNameLookup = func() map[string]tcell.Key {
	lookup := make(map[string]tcell.Key, len(tcell.KeyNames)+8)
	for key, name := range tcell.KeyNames {
		lookup[strings.ToLower(name)] = key
	}

	lookup["return"] = tcell.KeyEnter
	lookup["pageup"] = tcell.KeyPgUp
	lookup["pagedown"] = tcell.KeyPgDn
	lookup["pgdown"] = tcell.KeyPgDn
	lookup["shift-tab"] = tcell.KeyBacktab
	lookup["ctrl-/"] = tcell.KeyCtrlUnderscore
	lookup["ctrl-?"] = tcell.KeyCtrlUnderscore
	lookup["ctrl-_"] = tcell.KeyCtrlUnderscore

	return lookup
}()
