package overlay

import (
	"strings"

	"github.com/rivo/tview"
)

const JQCheatSheet = `[blue::b]Basics[-:-:-]
  [yellow].[-]                    identity (return input)
  [yellow].foo[-]                 object field
  [yellow].[0][-]                 array element by index
  [yellow].[][-]                  iterate array/object values
  [yellow].foo?[-]                optional lookup (no error)

[blue::b]Selection / filtering[-:-:-]
  [yellow]map(select(.x > 0))[-]  keep matching elements
  [yellow].[] | select(.ok)[-]    stream matching values
  [yellow]any(.[]; .ok)[-]        true if any item matches
  [yellow]all(.[]; .ok)[-]        true if all items match

[blue::b]Object / array building[-:-:-]
  [yellow]{id: .id, name: .n}[-]  build object
  [yellow][.items[].id][-]        build array from stream
  [yellow]with_entries(...)[-]    transform object keys/values

[blue::b]Common transforms[-:-:-]
  [yellow]sort_by(.ts)[-]         sort by field
  [yellow]group_by(.type)[-]      group sorted items
  [yellow]unique[-]               unique values
  [yellow]add[-]                  sum/merge array values
  [yellow]length[-]               length of array/string/object

[blue::b]Strings and formatting[-:-:-]
  [yellow]tostring[-]             JSON -> string
  [yellow]tonumber[-]             string -> number
  [yellow]@json[-]                escape as JSON string
  [yellow]@csv[-]                 format as CSV row

[blue::b]Tip[-:-:-]
  Compose small filters with [yellow]|[-] and test incrementally.`

var cheatSheetSize = func() struct {
	Width  int
	Height int
} {
	lines := strings.Split(JQCheatSheet, "\n")
	maxWidth := 0
	for _, line := range lines {
		if width := tview.TaggedStringWidth(line); width > maxWidth {
			maxWidth = width
		}
	}

	return struct {
		Width  int
		Height int
	}{Width: maxWidth, Height: len(lines) + 1}
}()
