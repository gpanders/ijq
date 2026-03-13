package overlay

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"codeberg.org/gpanders/ijq/internal/options"
)

const (
	rootMenuPage      = "overlay-root-menu"
	configurePage     = "overlay-configure"
	historyPage       = "overlay-history"
	confirmDeletePage = "overlay-confirm-delete"
	cheatSheetPage    = "overlay-cheat-sheet"
	keybindingsPage   = "overlay-keybindings"

	smallWidth    = 50
	menuHeight    = 10
	historyHeight = 15
)

type mode int

const (
	modeRoot mode = iota
	modeConfigure
	modeHistoryList
	modeHistoryFilter
	modeHistoryConfirmDelete
	modeCheatSheet
	modeKeybindings
)

func (m mode) IsTextInput() bool {
	switch m {
	case modeHistoryFilter:
		return true
	default:
		return false
	}
}

const (
	historyHelpText    = "[::d]Enter[::-] [::b]select[::-]   [::d]/[::-] [::b]filter[::-]   [::d]X[::-] [::b]delete[::-]"
	rootHelpText       = "[::d]Esc/Ctrl-C[::-] [::b]close[::-]   [::d]Space/Enter[::-] [::b]select[::-]"
	configureHelpText  = "[::d]Space/Enter[::-] [::b]toggle[::-]"
	cheatSheetHelpText = "[::d]Esc/Ctrl-C[::-] [::b]close[::-]"
	keybindHelpText    = "[::d]Esc/Ctrl-C[::-] [::b]close[::-]"

	confirmDeletePromptText = "Delete the following entry from history?"
	confirmDeleteHeight     = 5
)

type KeybindingEntry struct {
	Action     string
	Keybinding string
}

type Callbacks struct {
	ConfigureRows              func() []string
	ToggleConfigureRow         func(option options.Option)
	SaveCurrentFilterToHistory func() (status string, err error)
	LoadHistoryEntries         func() []string
	DeleteHistoryEntryAt       func(index int) error
	ApplyHistoryEntry          func(expr string)
	ActiveKeybindings          func() []KeybindingEntry
	OpenFocusedPaneInEditor    func()
}

type Controller struct {
	app      *tview.Application
	pages    *tview.Pages
	pageName string

	callbacks Callbacks

	container *tview.Grid
	subpages  *tview.Pages

	rootMenu   *tview.List
	configure  *tview.List
	history    *tview.List
	cheatSheet *tview.TextView
	keybinds   *tview.TextView

	rootLayout       *tview.Flex
	configureLayout  *tview.Flex
	historyLayout    *tview.Flex
	cheatSheetLayout *tview.Flex
	keybindsLayout   *tview.Flex

	rootHelpTextView       *tview.TextView
	configureHelpTextView  *tview.TextView
	historyHelpTextView    *tview.TextView
	cheatSheetHelpTextView *tview.TextView
	keybindHelpTextView    *tview.TextView

	historyFilterInput *tview.InputField
	confirmDeleteView  *tview.TextView

	open          bool
	mode          mode
	previousFocus tview.Primitive

	historyEntries         []string
	historyFilteredIndexes []int
	historyQuery           string
	historyQueryBeforeEdit string
	historyFilterVisible   bool

	pendingDeleteIndex int
	pendingDeleteEntry string
	confirmDeleteYes   bool
}

func NewController(app *tview.Application, pages *tview.Pages, pageName string, callbacks Callbacks) *Controller {
	c := &Controller{
		app:                app,
		pages:              pages,
		pageName:           pageName,
		callbacks:          callbacks,
		pendingDeleteIndex: -1,
	}

	c.rootMenu = newList("Menu")
	c.rootMenu.AddItem("Configure", "", 0, nil)
	c.rootMenu.AddItem("Save current filter to history", "", 0, nil)
	c.rootMenu.AddItem("Manage history", "", 0, nil)
	c.rootMenu.AddItem("Open focused pane in editor", "", 0, nil)
	c.rootMenu.AddItem("Keybindings", "", 0, nil)
	c.rootMenu.AddItem("Cheat sheet", "", 0, nil)

	c.configure = newList("Configure")
	c.configure.SetUseStyleTags(true, false)
	c.configure.SetTitle("Configure")
	c.configure.SetBorderPadding(0, 0, 1, 1)

	c.history = newList("History")

	c.historyFilterInput = tview.NewInputField()
	c.historyFilterInput.SetFieldBackgroundColor(tcell.ColorDefault)
	c.historyFilterInput.SetFieldTextColor(tcell.ColorDefault)
	c.historyFilterInput.SetLabel("Filter: ")
	c.historyFilterInput.SetChangedFunc(func(text string) {
		c.historyQuery = text
		c.refreshHistory(c.history.GetCurrentItem())
	})

	c.rootHelpTextView = tview.NewTextView()
	c.rootHelpTextView.SetDynamicColors(true)
	c.rootHelpTextView.SetTextAlign(tview.AlignCenter)
	c.rootHelpTextView.SetText(rootHelpText)

	c.configureHelpTextView = tview.NewTextView()
	c.configureHelpTextView.SetDynamicColors(true)
	c.configureHelpTextView.SetTextAlign(tview.AlignCenter)
	c.configureHelpTextView.SetText(configureHelpText)

	c.historyHelpTextView = tview.NewTextView()
	c.historyHelpTextView.SetDynamicColors(true)
	c.historyHelpTextView.SetTextAlign(tview.AlignCenter)
	c.historyHelpTextView.SetText(historyHelpText)

	c.confirmDeleteView = tview.NewTextView()
	c.confirmDeleteView.SetBorder(true)
	c.confirmDeleteView.SetTitle("Confirm delete")
	c.confirmDeleteView.SetWrap(true)
	c.confirmDeleteView.SetTextAlign(tview.AlignCenter)
	c.confirmDeleteView.SetDynamicColors(true)
	c.confirmDeleteView.SetBorderPadding(0, 0, 1, 1)

	c.cheatSheet = tview.NewTextView()
	c.cheatSheet.SetBorder(true)
	c.cheatSheet.SetTitle("jq cheat sheet")
	c.cheatSheet.SetWrap(false)
	c.cheatSheet.SetDynamicColors(true)
	c.cheatSheet.SetText(JQCheatSheet)
	c.cheatSheet.SetBorderPadding(0, 0, 1, 1)

	c.keybinds = tview.NewTextView()
	c.keybinds.SetBorder(true)
	c.keybinds.SetTitle("Keybindings")
	c.keybinds.SetWrap(false)
	c.keybinds.SetDynamicColors(true)
	c.keybinds.SetBorderPadding(0, 0, 1, 1)

	c.cheatSheetHelpTextView = tview.NewTextView()
	c.cheatSheetHelpTextView.SetDynamicColors(true)
	c.cheatSheetHelpTextView.SetTextAlign(tview.AlignCenter)
	c.cheatSheetHelpTextView.SetText(cheatSheetHelpText)

	c.keybindHelpTextView = tview.NewTextView()
	c.keybindHelpTextView.SetDynamicColors(true)
	c.keybindHelpTextView.SetTextAlign(tview.AlignCenter)
	c.keybindHelpTextView.SetText(keybindHelpText)

	c.rootLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.rootMenu, 0, 1, true).
		AddItem(c.rootHelpTextView, 1, 0, false)

	c.configureLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.configure, 0, 1, true).
		AddItem(c.configureHelpTextView, 1, 0, false)

	c.historyLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.history, 0, 1, true).
		AddItem(c.historyFilterInput, 0, 0, false).
		AddItem(c.historyHelpTextView, 1, 0, false)

	c.cheatSheetLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.cheatSheet, 0, 1, true).
		AddItem(c.cheatSheetHelpTextView, 1, 0, false)

	c.keybindsLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.keybinds, 0, 1, true).
		AddItem(c.keybindHelpTextView, 1, 0, false)

	c.subpages = tview.NewPages().
		AddPage(rootMenuPage, c.rootLayout, true, true).
		AddPage(configurePage, c.configureLayout, true, false).
		AddPage(historyPage, c.historyLayout, true, false).
		AddPage(confirmDeletePage, c.confirmDeleteView, true, false).
		AddPage(keybindingsPage, c.keybindsLayout, true, false).
		AddPage(cheatSheetPage, c.cheatSheetLayout, true, false)

	c.container = tview.NewGrid().
		SetRows(0, menuHeight, 0).
		SetColumns(0, smallWidth, 0).
		AddItem(c.subpages, 1, 1, 1, 1, 0, 0, true)

	return c
}

func (c *Controller) Primitive() tview.Primitive {
	return c.container
}

func (c *Controller) IsOpen() bool {
	return c.open
}

func (c *Controller) Open() {
	if c.open {
		return
	}

	c.previousFocus = c.app.GetFocus()
	c.pages.ShowPage(c.pageName)
	c.pages.SendToFront(c.pageName)
	c.open = true
	c.showRootMenu("")
}

func (c *Controller) Close() {
	if !c.open {
		return
	}

	c.pages.HidePage(c.pageName)
	c.open = false

	if c.previousFocus != nil {
		c.app.SetFocus(c.previousFocus)
	}
}

func (c *Controller) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if !c.open {
		return event
	}

	// Navigation keymaps that apply to all (non-text input) modes
	if !c.mode.IsTextInput() {
		switch event.Key() {
		case tcell.KeyRune:
			if event.Modifiers() == tcell.ModNone {
				switch event.Rune() {
				case 'j':
					return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
				case 'k':
					return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
				case 'q':
					c.Close()
					return nil
				}
			}
		}
	}

	// Mode specific handling
	switch c.mode {
	case modeRoot:
		switch event.Key() {
		case tcell.KeyRune:
			if event.Modifiers() == tcell.ModNone {
				switch event.Rune() {
				case ' ':
					c.activateRootMenu(c.rootMenu.GetCurrentItem())
					return nil
				}
			}
		case tcell.KeyEnter:
			c.activateRootMenu(c.rootMenu.GetCurrentItem())
			return nil
		case tcell.KeyCtrlC, tcell.KeyEsc:
			c.Close()
			return nil
		}
	case modeConfigure:
		switch event.Key() {
		case tcell.KeyRune:
			if event.Modifiers() == tcell.ModNone {
				switch event.Rune() {
				case ' ':
					c.toggleConfigure(c.configure.GetCurrentItem())
					return nil
				}
			}
		case tcell.KeyEnter:
			c.toggleConfigure(c.configure.GetCurrentItem())
			return nil
		case tcell.KeyCtrlC, tcell.KeyEsc:
			c.showRootMenu("")
			return nil
		}
	case modeHistoryConfirmDelete:
		switch event.Key() {
		case tcell.KeyEnter:
			c.confirmDeleteSelection(c.confirmDeleteYes)
			return nil
		case tcell.KeyEsc, tcell.KeyCtrlC:
			c.confirmDeleteSelection(false)
			return nil

		case tcell.KeyLeft, tcell.KeyBacktab:
			c.confirmDeleteYes = true
			c.renderConfirmDeletePrompt()
			return nil

		case tcell.KeyRight, tcell.KeyTab:
			c.confirmDeleteYes = false
			c.renderConfirmDeletePrompt()
			return nil

		case tcell.KeyRune:
			if event.Modifiers() == tcell.ModNone {
				switch event.Rune() {
				case 'h':
					c.confirmDeleteYes = true
					c.renderConfirmDeletePrompt()
					return nil
				case 'l':
					c.confirmDeleteYes = false
					c.renderConfirmDeletePrompt()
					return nil
				}
			}
		default:
			return nil
		}
	case modeHistoryFilter:
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEsc:
			c.endHistoryFilterEdit(true)
			return nil
		case tcell.KeyEnter:
			c.endHistoryFilterEdit(false)
			return nil
		}
	case modeHistoryList:
		switch event.Key() {
		case tcell.KeyRune:
			if event.Modifiers() == tcell.ModNone {
				switch event.Rune() {
				case '/':
					c.beginHistoryFilterEdit()
					return nil
				case 'x', 'X':
					c.promptDeleteSelectedHistoryEntry()
					return nil
				}
			}
		case tcell.KeyEnter:
			c.applySelectedHistoryEntry()
			return nil
		case tcell.KeyCtrlC, tcell.KeyEsc:
			c.showRootMenu("")
			return nil
		}
	case modeCheatSheet, modeKeybindings:
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEsc:
			c.showRootMenu("")
			return nil
		}
	}

	return event
}

func (c *Controller) showRootMenu(status string) {
	c.mode = modeRoot
	c.subpages.SwitchToPage(rootMenuPage)
	c.resize(smallWidth, menuHeight)
	c.setHistoryFilterVisible(false)
	if status == "" {
		c.rootMenu.SetTitle("Menu")
	} else {
		c.rootMenu.SetTitle(fmt.Sprintf("Menu (%s)", status))
	}

	c.app.SetFocus(c.rootMenu)
}

func (c *Controller) activateRootMenu(index int) {
	switch index {
	case 0:
		c.showConfigure()
	case 1:
		c.saveCurrentFilterToHistory()
	case 2:
		c.showHistory()
	case 3:
		c.Close()
		c.callbacks.OpenFocusedPaneInEditor()
	case 4:
		c.showKeybindings()
	case 5:
		c.showCheatSheet()
	}
}

func (c *Controller) showConfigure() {
	c.mode = modeConfigure
	c.subpages.SwitchToPage(configurePage)
	c.refreshConfigure(c.configure.GetCurrentItem())
	c.resize(configureSize.Width, configureSize.Height)
	c.app.SetFocus(c.configure)
}

func (c *Controller) refreshConfigure(current int) []string {
	rows := []string{}
	if c.callbacks.ConfigureRows != nil {
		rows = c.callbacks.ConfigureRows()
	}

	c.configure.Clear()
	for _, row := range rows {
		c.configure.AddItem(" "+row+" ", "", 0, nil)
	}

	if len(rows) == 0 {
		return rows
	}

	if current < 0 {
		current = 0
	}

	if current >= len(rows) {
		current = len(rows) - 1
	}

	c.configure.SetCurrentItem(current)

	return rows
}

func (c *Controller) toggleConfigure(index int) {
	if c.callbacks.ToggleConfigureRow == nil {
		return
	}

	if index < 0 || index >= len(configureRows) {
		return
	}

	c.callbacks.ToggleConfigureRow(configureRows[index])
	c.refreshConfigure(index)
}

func (c *Controller) saveCurrentFilterToHistory() {
	if c.callbacks.SaveCurrentFilterToHistory == nil {
		c.showRootMenu("save action unavailable")
		return
	}

	status, err := c.callbacks.SaveCurrentFilterToHistory()
	if err != nil {
		c.showRootMenu(err.Error())
		return
	}

	c.showRootMenu(status)
}

func (c *Controller) showHistory() {
	c.mode = modeHistoryList
	c.subpages.SwitchToPage(historyPage)
	c.resize(smallWidth, historyHeight)
	c.setHistoryFilterVisible(false)
	c.historyQuery = ""
	c.historyQueryBeforeEdit = ""
	c.historyFilterInput.SetText("")
	if c.callbacks.LoadHistoryEntries != nil {
		c.historyEntries = c.callbacks.LoadHistoryEntries()
	} else {
		c.historyEntries = nil
	}

	c.refreshHistory(0)
	c.app.SetFocus(c.history)
}

func (c *Controller) refreshHistory(current int) {
	c.historyFilteredIndexes = filterIndexes(c.historyEntries, c.historyQuery)

	c.history.Clear()
	for _, index := range c.historyFilteredIndexes {
		c.history.AddItem(c.historyEntries[index], "", 0, nil)
	}

	if len(c.historyFilteredIndexes) > 0 {
		if current < 0 {
			current = 0
		}

		if current >= len(c.historyFilteredIndexes) {
			current = len(c.historyFilteredIndexes) - 1
		}

		c.history.SetCurrentItem(current)
	}

	c.updateHistoryTitle("")
}

func (c *Controller) updateHistoryTitle(status string) {
	title := fmt.Sprintf("History (%s)", formatCount(len(c.historyFilteredIndexes), len(c.historyEntries)))
	if strings.TrimSpace(status) != "" {
		title = fmt.Sprintf("%s - %s", title, status)
	}

	c.history.SetTitle(title)
}

func (c *Controller) beginHistoryFilterEdit() {
	c.historyQueryBeforeEdit = c.historyQuery
	c.mode = modeHistoryFilter
	c.setHistoryFilterVisible(true)
	c.app.SetFocus(c.historyFilterInput)
}

func (c *Controller) endHistoryFilterEdit(cancel bool) {
	if cancel {
		c.historyQuery = c.historyQueryBeforeEdit
		c.historyFilterInput.SetText(c.historyQuery)
		c.refreshHistory(c.history.GetCurrentItem())
	}

	c.mode = modeHistoryList
	c.setHistoryFilterVisible(strings.TrimSpace(c.historyQuery) != "")
	c.app.SetFocus(c.history)
}

func (c *Controller) setHistoryFilterVisible(visible bool) {
	c.historyFilterVisible = visible
	if visible {
		c.historyLayout.ResizeItem(c.historyFilterInput, 1, 0)
		return
	}

	c.historyLayout.ResizeItem(c.historyFilterInput, 0, 0)
}

func (c *Controller) promptDeleteSelectedHistoryEntry() {
	if c.callbacks.DeleteHistoryEntryAt == nil || len(c.historyFilteredIndexes) == 0 {
		return
	}

	selected := c.history.GetCurrentItem()
	if selected < 0 || selected >= len(c.historyFilteredIndexes) {
		return
	}

	index := c.historyFilteredIndexes[selected]
	entry := c.historyEntries[index]

	c.pendingDeleteIndex = index
	c.pendingDeleteEntry = entry
	c.confirmDeleteYes = true
	c.mode = modeHistoryConfirmDelete
	c.subpages.SwitchToPage(confirmDeletePage)

	width := max(
		tview.TaggedStringWidth(tview.Escape(strings.ReplaceAll(entry, "\n", " "))),
		tview.TaggedStringWidth(confirmDeletePromptText),
	)

	c.resize(width, confirmDeleteHeight)
	c.renderConfirmDeletePrompt()
	c.app.SetFocus(c.confirmDeleteView)
}

func (c *Controller) confirmDeleteSelection(yes bool) {
	index := c.pendingDeleteIndex
	c.pendingDeleteIndex = -1
	c.pendingDeleteEntry = ""
	c.confirmDeleteYes = true

	c.mode = modeHistoryList
	c.subpages.SwitchToPage(historyPage)
	c.resize(smallWidth, historyHeight)
	c.app.SetFocus(c.history)

	if !yes || index < 0 {
		c.updateHistoryTitle("")
		return
	}

	if err := c.callbacks.DeleteHistoryEntryAt(index); err != nil {
		c.updateHistoryTitle(err.Error())
		return
	}

	if c.callbacks.LoadHistoryEntries != nil {
		c.historyEntries = c.callbacks.LoadHistoryEntries()
	}

	c.refreshHistory(c.history.GetCurrentItem())
	c.updateHistoryTitle("deleted")
}

func (c *Controller) renderConfirmDeletePrompt() {
	entry := tview.Escape(strings.ReplaceAll(c.pendingDeleteEntry, "\n", " "))

	yes := " Yes "
	no := " No "
	if c.confirmDeleteYes {
		yes = "[::r] Yes [-:-:-]"
	} else {
		no = "[::r] No [-:-:-]"
	}

	c.confirmDeleteView.SetText(fmt.Sprintf("%s\n\n[yellow]%s[-]\n\n%s %s", confirmDeletePromptText, entry, yes, no))
}

func (c *Controller) applySelectedHistoryEntry() {
	if c.callbacks.ApplyHistoryEntry == nil || len(c.historyFilteredIndexes) == 0 {
		return
	}

	selected := c.history.GetCurrentItem()
	if selected < 0 || selected >= len(c.historyFilteredIndexes) {
		return
	}

	entry := c.historyEntries[c.historyFilteredIndexes[selected]]
	c.callbacks.ApplyHistoryEntry(entry)
	c.Close()
}

func (c *Controller) showCheatSheet() {
	c.mode = modeCheatSheet
	c.subpages.SwitchToPage(cheatSheetPage)
	c.app.SetFocus(c.cheatSheet)
	c.resize(cheatSheetSize.Width, cheatSheetSize.Height)
}

func (c *Controller) showKeybindings() {
	c.mode = modeKeybindings
	c.subpages.SwitchToPage(keybindingsPage)
	c.app.SetFocus(c.keybinds)

	entries := []KeybindingEntry{}
	if c.callbacks.ActiveKeybindings != nil {
		entries = c.callbacks.ActiveKeybindings()
	}

	content, width, height := formatKeybindingRows(entries)
	c.keybinds.SetText(content)
	c.resize(width, height)
}

func formatKeybindingRows(entries []KeybindingEntry) (content string, width int, height int) {
	if len(entries) == 0 {
		line := "No active keybindings"
		return line, len(line) + 2, 4
	}

	maxActionWidth := 0
	maxBindingWidth := 0
	for _, entry := range entries {
		if actionWidth := tview.TaggedStringWidth(entry.Action); actionWidth > maxActionWidth {
			maxActionWidth = actionWidth
		}
		if bindingWidth := tview.TaggedStringWidth(entry.Keybinding); bindingWidth > maxBindingWidth {
			maxBindingWidth = bindingWidth
		}
	}

	lines := make([]string, 0, len(entries))
	const gap = 3
	for _, entry := range entries {
		actionWidth := tview.TaggedStringWidth(entry.Action)
		bindingWidth := tview.TaggedStringWidth(entry.Keybinding)
		spaces := max((maxActionWidth-actionWidth)+gap+(maxBindingWidth-bindingWidth), gap)

		lines = append(lines, fmt.Sprintf("%s%s[::b]%s[-:-:-]", entry.Action, strings.Repeat(" ", spaces), entry.Keybinding))
	}

	contentWidth := maxActionWidth + gap + maxBindingWidth
	return strings.Join(lines, "\n"), contentWidth, len(lines) + 1
}

func (c *Controller) resize(width int, height int) {
	// Add 2 to height to account for border
	c.container.SetRows(0, height+2, 0)
	// Add 4 to height to account for border and padding
	c.container.SetColumns(0, width+4, 0)
}

func newList(title string) *tview.List {
	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetWrapAround(false)
	list.SetUseStyleTags(false, false)
	list.SetHighlightFullLine(true)
	list.SetSelectedFocusOnly(false)
	list.SetSelectedStyle(tcell.StyleDefault.Reverse(true))
	list.SetBorder(true)
	list.SetTitle(title)
	list.SetBorderPadding(0, 0, 1, 1)

	return list
}
