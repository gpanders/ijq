package overlay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterIndexesEmptyQueryReturnsAll(t *testing.T) {
	entries := []string{".foo", ".bar", ".baz"}

	indexes := filterIndexes(entries, "")
	assert.Equal(t, []int{0, 1, 2}, indexes)
}

func TestFilterIndexesCaseInsensitiveSubstring(t *testing.T) {
	entries := []string{".foo", ".Bar", ".baz", ".Foobar"}

	indexes := filterIndexes(entries, "foo")
	assert.Equal(t, []int{0, 3}, indexes)
}

func TestFormatCount(t *testing.T) {
	assert.Equal(t, "showing 3 of 12 entries", formatCount(3, 12))
	assert.Equal(t, "showing 0 of 0 entries", formatCount(-1, -1))
}
