package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultKeymap(t *testing.T) {
	t.Parallel()

	keymap := DefaultKeymap()

	assert.True(t, keymap.SubmitFilter.Matches(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)))
	assert.True(t, keymap.ToggleInputPane.Matches(tcell.NewEventKey(tcell.KeyCtrlO, ' ', tcell.ModNone)))
	assert.True(t, keymap.SaveFilterHistory.Matches(tcell.NewEventKey(tcell.KeyCtrlS, ' ', tcell.ModNone)))
	assert.True(t, keymap.ToggleMenu.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModNone)))
	assert.True(t, keymap.ToggleMenu.Matches(tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModCtrl)))
	assert.True(t, keymap.ToggleMenu.Matches(tcell.NewEventKey(tcell.KeyRune, '_', tcell.ModCtrl)))
	assert.Empty(t, keymap.CopyFilterToClipboard)
	assert.Empty(t, keymap.CopyOutputToClipboard)
	assert.True(t, keymap.PageUp.Matches(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone)))
	assert.True(t, keymap.PageUp.Matches(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModAlt)))
	assert.True(t, keymap.LineStart.Matches(tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone)))
}

func TestKeymapEntries(t *testing.T) {
	t.Parallel()

	keymap := Keymap{
		MoveDown:             KeyBindings{{key: tcell.KeyRune, rune: 'j'}},
		PageDown:             KeyBindings{{key: tcell.KeyRune, rune: 'f'}, {key: tcell.KeyRune, rune: 'F'}},
		SubmitFilter:         KeyBindings{{key: tcell.KeyEnter}},
		NextAutocomplete:     KeyBindings{{key: tcell.KeyTab}},
		PreviousAutocomplete: KeyBindings{{key: tcell.KeyBacktab}},
		ToggleMenu:           KeyBindings{{key: tcell.KeyCtrlUnderscore}},
	}

	assert.Equal(t, []KeymapEntry{
		{Action: "move-down", Keybinding: "j"},
		{Action: "next-autocomplete", Keybinding: "Tab"},
		{Action: "page-down", Keybinding: "f, F"},
		{Action: "previous-autocomplete", Keybinding: "Shift-Tab"},
		{Action: "submit-filter", Keybinding: "Enter"},
		{Action: "toggle-menu", Keybinding: "Ctrl-/"},
	}, keymap.Entries())
}

func TestKeyBindingString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ctrl-/", KeyBinding{key: tcell.KeyCtrlUnderscore}.String())
	assert.Equal(t, "Ctrl-q", KeyBinding{key: tcell.KeyRune, rune: 'q', mods: tcell.ModCtrl}.String())
	assert.Equal(t, "Enter", KeyBinding{key: tcell.KeyEnter}.String())
}

func TestUnmarshalTextAltRune(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Alt+v"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModAlt)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone)))
}

func TestUnmarshalTextCtrlPlusLetter(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl+N"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlN, ' ', tcell.ModNone)))
}

func TestUnmarshalTextCtrlHyphenLetter(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl-N"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlN, ' ', tcell.ModNone)))
}

func TestUnmarshalTextShiftHyphenArrow(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Shift-Up"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModShift)))
}

func TestUnmarshalTextAltShiftLetter(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Alt+Shift+a"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModAlt)))
}

func TestUnmarshalTextShiftLetter(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Shift+g"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone)))
}

func TestRuneBindingDoesNotMatchModifiedRune(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("u"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt)))
}

func TestUnmarshalTextPlusRune(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("+"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, '+', tcell.ModNone)))
}

func TestUnmarshalTextCtrlPlusRune(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl++"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyRune, '+', tcell.ModCtrl)))
}

func TestUnmarshalTextCtrlSlashAlias(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl+/"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModNone)))
}

func TestUnmarshalTextCtrlQuestionAlias(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl-?"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModNone)))
}

func TestUnmarshalTextCtrlUnderscoreAlias(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl-_"))
	assert.NoError(t, err)

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModNone)))
}

func TestUnmarshalTextRejectsEmptyValue(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte(" "))
	assert.Error(t, err)
}

func TestUnmarshalTextRejectsUnknownModifier(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Hyper+a"))
	assert.Error(t, err)
}

func TestUnmarshalTextRejectsModifierWithoutKey(t *testing.T) {
	t.Parallel()

	var binding KeyBinding
	err := binding.UnmarshalText([]byte("Ctrl+"))
	assert.Error(t, err)
}

func TestNonRuneBindingDoesNotMatchExtraModifier(t *testing.T) {
	t.Parallel()

	binding := KeyBinding{key: tcell.KeyEnter}

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModCtrl)))
}

func TestBindingRequiresExactModifiers(t *testing.T) {
	t.Parallel()

	binding := KeyBinding{key: tcell.KeyUp, mods: tcell.ModShift}

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModShift)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModShift|tcell.ModCtrl)))
}

func TestControlKeyBindingMatchesCtrlModifier(t *testing.T) {
	t.Parallel()

	binding := KeyBinding{key: tcell.KeyCtrlUnderscore}

	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModCtrl)))
	assert.True(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModCtrl|tcell.ModShift)))
	assert.False(t, binding.Matches(tcell.NewEventKey(tcell.KeyCtrlUnderscore, ' ', tcell.ModAlt)))
}
