package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsageHidesLegacyFlags(t *testing.T) {
	options := Options{}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &options, &out)
	flagSet.Usage()

	help := out.String()
	assert.Contains(t, help, "-c")
	assert.Contains(t, help, "-f")
	assert.NotContains(t, help, "-H")
	assert.NotContains(t, help, "-jqbin")
	assert.NotContains(t, help, "-hide-input-pane")
}

func TestLegacyFlagsRemainSupported(t *testing.T) {
	options := Options{}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &options, &out)
	err := flagSet.Parse([]string{"-H", "", "-jqbin", "custom-jq", "-hide-input-pane"})
	assert.NoError(t, err)

	assert.Equal(t, "", options.config.HistoryFile)
	assert.Equal(t, "custom-jq", options.config.JQCommand)
	assert.True(t, options.config.HideInputPane)
}

func TestLibraryPathsFromConfigAndFlags(t *testing.T) {
	options := Options{
		config: Config{
			LibraryPaths: LibraryPaths{"/config/modules"},
		},
	}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &options, &out)
	err := flagSet.Parse([]string{"-L", "/cli/modules"})
	assert.NoError(t, err)

	assert.Equal(t, LibraryPaths{"/config/modules", "/cli/modules"}, options.config.LibraryPaths)
}

func TestLegacyFlagsOverrideConfigValues(t *testing.T) {
	options := Options{
		config: Config{
			HistoryFile:   "/config/history",
			JQCommand:     "/usr/bin/jq",
			HideInputPane: false,
		},
	}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &options, &out)
	err := flagSet.Parse([]string{"-H", "", "-jqbin", "custom-jq", "-hide-input-pane"})
	assert.NoError(t, err)

	assert.Equal(t, "", options.config.HistoryFile)
	assert.Equal(t, "custom-jq", options.config.JQCommand)
	assert.True(t, options.config.HideInputPane)
}
