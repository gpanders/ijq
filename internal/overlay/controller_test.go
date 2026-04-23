package overlay

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"

	"codeberg.org/gpanders/ijq/internal/options"
)

func newOpenController(t *testing.T, callbacks Callbacks) *Controller {
	t.Helper()

	app := tview.NewApplication()
	pages := tview.NewPages()
	controller := NewController(app, pages, "overlay", callbacks)
	pages.AddPage("overlay", controller.Primitive(), true, false)
	controller.Open()

	return controller
}

func keyEvent(key tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(key, ' ', tcell.ModNone)
}

func runeEvent(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

func TestHandleInputReturnsEventWhenClosed(t *testing.T) {
	t.Parallel()

	controller := NewController(tview.NewApplication(), tview.NewPages(), "overlay", Callbacks{})
	event := keyEvent(tcell.KeyEnter)

	assert.Equal(t, event, controller.HandleInput(event))
}

func TestHandleInputGlobalNavigationAndClose(t *testing.T) {
	t.Parallel()

	controller := newOpenController(t, Callbacks{})

	event := controller.HandleInput(runeEvent('j'))
	assert.Equal(t, tcell.KeyDown, event.Key())

	event = controller.HandleInput(runeEvent('k'))
	assert.Equal(t, tcell.KeyUp, event.Key())

	event = controller.HandleInput(runeEvent('q'))
	assert.Nil(t, event)
	assert.False(t, controller.IsOpen())
}

func TestHandleInputConfigureModeTogglesAndReturnsRoot(t *testing.T) {
	t.Parallel()

	var toggledOption options.Option

	controller := newOpenController(t, Callbacks{
		ConfigureRows: func() []string {
			return []string{"Option"}
		},
		ToggleConfigureRow: func(option options.Option) {
			toggledOption = option
		},
	})

	controller.rootMenu.SetCurrentItem(0)
	event := controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.Equal(t, modeConfigure, controller.mode)

	event = controller.HandleInput(runeEvent(' '))
	assert.Nil(t, event)
	if assert.NotNil(t, toggledOption) {
		assert.Equal(t, "c", toggledOption.Flag())
	}

	event = controller.HandleInput(keyEvent(tcell.KeyEsc))
	assert.Nil(t, event)
	assert.Equal(t, modeRoot, controller.mode)
}

func TestHandleInputHistoryFilterCancelRestoresQuery(t *testing.T) {
	t.Parallel()

	controller := newOpenController(t, Callbacks{
		LoadHistoryEntries: func() []string {
			return []string{".foo", ".bar"}
		},
	})

	controller.rootMenu.SetCurrentItem(4)
	controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Equal(t, modeHistoryList, controller.mode)

	controller.historyQuery = "foo"
	controller.historyFilterInput.SetText("foo")

	event := controller.HandleInput(runeEvent('/'))
	assert.Nil(t, event)
	assert.Equal(t, modeHistoryFilter, controller.mode)
	assert.True(t, controller.historyFilterVisible)

	controller.historyFilterInput.SetText("bar")
	controller.HandleInput(keyEvent(tcell.KeyEsc))
	assert.Equal(t, modeHistoryList, controller.mode)
	assert.Equal(t, "foo", controller.historyQuery)
	assert.True(t, controller.historyFilterVisible)
}

func TestHandleInputHistoryFilterAllowsTypingNavigationKeys(t *testing.T) {
	t.Parallel()

	controller := newOpenController(t, Callbacks{
		LoadHistoryEntries: func() []string {
			return []string{".foo", ".bar"}
		},
	})

	controller.rootMenu.SetCurrentItem(4)
	controller.HandleInput(keyEvent(tcell.KeyEnter))
	controller.HandleInput(runeEvent('/'))
	assert.Equal(t, modeHistoryFilter, controller.mode)

	event := controller.HandleInput(runeEvent('j'))
	if assert.NotNil(t, event) {
		assert.Equal(t, tcell.KeyRune, event.Key())
		assert.Equal(t, 'j', event.Rune())
	}

	event = controller.HandleInput(runeEvent('k'))
	if assert.NotNil(t, event) {
		assert.Equal(t, tcell.KeyRune, event.Key())
		assert.Equal(t, 'k', event.Rune())
	}

	event = controller.HandleInput(runeEvent('q'))
	if assert.NotNil(t, event) {
		assert.Equal(t, tcell.KeyRune, event.Key())
		assert.Equal(t, 'q', event.Rune())
	}

	assert.True(t, controller.IsOpen())
}

func TestHandleInputHistoryFilterCancelEventIsConsumed(t *testing.T) {
	t.Parallel()

	controller := newOpenController(t, Callbacks{
		LoadHistoryEntries: func() []string {
			return []string{".foo", ".bar"}
		},
	})

	controller.rootMenu.SetCurrentItem(4)
	controller.HandleInput(keyEvent(tcell.KeyEnter))
	controller.HandleInput(runeEvent('/'))
	assert.Equal(t, modeHistoryFilter, controller.mode)

	event := controller.HandleInput(keyEvent(tcell.KeyEsc))
	assert.Nil(t, event)
	assert.Equal(t, modeHistoryList, controller.mode)
}

func TestHandleInputHistoryDeleteConfirmFlow(t *testing.T) {
	t.Parallel()

	entries := []string{"one", "two"}
	deletedIndex := -1

	controller := newOpenController(t, Callbacks{
		LoadHistoryEntries: func() []string {
			return append([]string(nil), entries...)
		},
		DeleteHistoryEntryAt: func(index int) error {
			deletedIndex = index
			entries = append(entries[:index], entries[index+1:]...)
			return nil
		},
	})

	controller.rootMenu.SetCurrentItem(4)
	controller.HandleInput(keyEvent(tcell.KeyEnter))
	controller.history.SetCurrentItem(1)

	event := controller.HandleInput(runeEvent('x'))
	assert.Nil(t, event)
	assert.Equal(t, modeHistoryConfirmDelete, controller.mode)
	assert.Equal(t, 1, controller.pendingDeleteIndex)

	event = controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.Equal(t, modeHistoryList, controller.mode)
	assert.Equal(t, 1, deletedIndex)
	assert.Equal(t, []string{"one"}, controller.historyEntries)
}

func TestHandleInputHistoryEnterAppliesEntryAndCloses(t *testing.T) {
	t.Parallel()

	appliedExpression := ""

	controller := newOpenController(t, Callbacks{
		LoadHistoryEntries: func() []string {
			return []string{".foo", ".bar"}
		},
		ApplyHistoryEntry: func(expression string) {
			appliedExpression = expression
		},
	})

	controller.rootMenu.SetCurrentItem(4)
	controller.HandleInput(keyEvent(tcell.KeyEnter))

	event := controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.Equal(t, ".foo", appliedExpression)
	assert.False(t, controller.IsOpen())
}

func TestHandleInputCopyActionsReturnToRoot(t *testing.T) {
	t.Parallel()

	copyFilterCalled := false
	copyOutputCalled := false

	controller := newOpenController(t, Callbacks{
		CopyFilterToClipboard: func() error {
			copyFilterCalled = true
			return nil
		},
		CopyOutputToClipboard: func() error {
			copyOutputCalled = true
			return nil
		},
	})

	controller.rootMenu.SetCurrentItem(2)
	event := controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.True(t, copyFilterCalled)
	assert.Equal(t, modeRoot, controller.mode)
	assert.Equal(t, "Menu (filter copied to clipboard)", controller.rootMenu.GetTitle())

	controller.rootMenu.SetCurrentItem(3)
	event = controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.True(t, copyOutputCalled)
	assert.Equal(t, modeRoot, controller.mode)
	assert.Equal(t, "Menu (output copied to clipboard)", controller.rootMenu.GetTitle())
}

func TestHandleInputCheatSheetAndKeybindingsReturnToRoot(t *testing.T) {
	t.Parallel()

	controller := newOpenController(t, Callbacks{
		ActiveKeybindings: func() []KeybindingEntry {
			return []KeybindingEntry{{Action: "submit", Keybinding: "Enter"}}
		},
	})

	controller.rootMenu.SetCurrentItem(6)
	event := controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.Equal(t, modeCheatSheet, controller.mode)

	event = controller.HandleInput(keyEvent(tcell.KeyEsc))
	assert.Nil(t, event)
	assert.Equal(t, modeRoot, controller.mode)

	controller.rootMenu.SetCurrentItem(5)
	event = controller.HandleInput(keyEvent(tcell.KeyEnter))
	assert.Nil(t, event)
	assert.Equal(t, modeKeybindings, controller.mode)

	event = controller.HandleInput(keyEvent(tcell.KeyCtrlC))
	assert.Nil(t, event)
	assert.Equal(t, modeRoot, controller.mode)
}
