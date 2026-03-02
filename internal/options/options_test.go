package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsToSliceReturnsEnabledFlagsInFieldOrder(t *testing.T) {
	opts := Options{
		CompactOutput: true,
		Slurp:         true,
		ASCIIOutput:   true,
		ForceColor:    true,
	}

	assert.Equal(t, []string{"-c", "-s", "-a", "-C"}, opts.ToSlice())
}

func TestOptionsToSliceReturnsEmptyWhenNoOptionsEnabled(t *testing.T) {
	assert.Empty(t, (Options{}).ToSlice())
}

func TestOptionsToSliceIncludesLibraryPaths(t *testing.T) {
	opts := Options{
		CompactOutput: true,
		LibraryPaths:  LibraryPaths{"foo", "bar"},
	}

	assert.Equal(t, []string{"-c", "-L", "foo", "-L", "bar"}, opts.ToSlice())
}

func TestOptionsToSliceExcludesNonRuntimeFlags(t *testing.T) {
	opts := Options{
		HideInputPane: true,
		JQCommand:     "custom-jq",
		HistoryFile:   "/tmp/history",
	}

	assert.Empty(t, opts.ToSlice())
}

func TestSetOnBoolOption(t *testing.T) {
	var compact CompactOutput

	assert.NoError(t, compact.Set("true"))
	assert.EqualValues(t, true, compact)

	assert.NoError(t, compact.Set("false"))
	assert.EqualValues(t, false, compact)

	assert.Error(t, compact.Set("not-a-bool"))
}

func TestSetOnLibraryPathsOption(t *testing.T) {
	var paths LibraryPaths

	assert.NoError(t, paths.Set("foo"))
	assert.NoError(t, paths.Set("bar"))
	assert.EqualValues(t, []string{"foo", "bar"}, paths)
}
