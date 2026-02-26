package overlay

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestConfigureRows(t *testing.T) {
	values := map[string]bool{"-R": true}

	rows := ConfigureRows(func(flag string) bool {
		return values[flag]
	})

	assert.Len(t, rows, len(configureOptions))
	assert.Equal(t, "● Read raw strings (-R)", rows[6])
	assert.Equal(t, "○ Force color (-C)", rows[8])
}

func TestToggleConfigureOptionForceColorDisablesMonochrome(t *testing.T) {
	values := map[string]bool{"-M": true}

	ok := ToggleConfigureOption(configureOptionIndex("-C"),
		func(flag string) bool {
			return values[flag]
		},
		func(flag string, value bool) {
			values[flag] = value
		},
	)

	assert.True(t, ok)
	assert.True(t, values["-C"])
	assert.False(t, values["-M"])
}

func TestToggleConfigureOptionMonochromeDisablesForceColor(t *testing.T) {
	values := map[string]bool{"-C": true}

	ok := ToggleConfigureOption(configureOptionIndex("-M"),
		func(flag string) bool {
			return values[flag]
		},
		func(flag string, value bool) {
			values[flag] = value
		},
	)

	assert.True(t, ok)
	assert.True(t, values["-M"])
	assert.False(t, values["-C"])
}

func TestToggleConfigureOptionRejectsInvalidIndex(t *testing.T) {
	values := map[string]bool{}

	assert.False(t, ToggleConfigureOption(-1,
		func(flag string) bool { return values[flag] },
		func(flag string, value bool) { values[flag] = value },
	))
	assert.False(t, ToggleConfigureOption(len(configureOptions),
		func(flag string) bool { return values[flag] },
		func(flag string, value bool) { values[flag] = value },
	))
}

func TestConfigureSizeMatchesRowWidth(t *testing.T) {
	rows := ConfigureRows(func(flag string) bool {
		return flag == "-n"
	})

	assert.Equal(t, len(rows)+3, configureSize.Height)

	maxWidth := 0
	for _, row := range rows {
		if rowWidth := tview.TaggedStringWidth(row); rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}

	assert.Equal(t, configureSize.Width-2, maxWidth)
}

func configureOptionIndex(flag string) int {
	for i, option := range configureOptions {
		if option.flag == flag {
			return i
		}
	}

	return -1
}
