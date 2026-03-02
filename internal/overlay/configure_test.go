package overlay

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureRowsContainsOnlyBoolOptionsInStructOrder(t *testing.T) {
	flags := make([]string, 0, len(configureRows))
	for _, option := range configureRows {
		value := reflect.Indirect(reflect.ValueOf(option))
		if assert.True(t, value.IsValid()) {
			assert.Equal(t, reflect.Bool, value.Kind())
		}

		flags = append(flags, option.Flag())
	}

	assert.Equal(t, []string{"c", "n", "s", "r", "j", "a", "R", "M", "C", "S", "hide-input-pane"}, flags)
}

func TestConfigureRowsExcludesNonBoolOptions(t *testing.T) {
	for _, option := range configureRows {
		assert.NotEqual(t, "L", option.Flag())
		assert.NotEqual(t, "jqbin", option.Flag())
		assert.NotEqual(t, "H", option.Flag())
	}
}

func TestConfigureSizeMatchesConfigureRows(t *testing.T) {
	maxWidth := len("Configure")
	for _, row := range configureRows {
		maxWidth = max(maxWidth, len(row.String()))
	}

	assert.Equal(t, maxWidth, configureSize.Width)
	assert.Equal(t, len(configureRows), configureSize.Height)
}
