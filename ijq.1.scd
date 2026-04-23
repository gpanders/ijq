ijq(1)

# NAME

ijq - interactive jq

# SYNOPSIS

*ijq* [*-cnsrjaRMCSV*] [*-f* _file_] [_filter_] [_files ..._]

# DESCRIPTION

*ijq* is a near drop-in replacement for *jq* that allows you to interactively
see the results of your filter as you construct it.

*ijq* contains two panes and an input field: the left pane is the original,
unmodified input data and the right pane contains the filtered output. When you
are finished, press Return to exit. The filtered output will be written to
standard output and the filter itself will be written to standard error.

*ijq* maintains a history of used filters unless disabled via
*history-file* in the config file. Delete all text in the filter field to
browse any available history.

If _files_ is omitted then *ijq* reads data from standard input.

An interactive menu is available with the *toggle-menu* action (default:
*Ctrl-/*, *Ctrl-?*, or *Ctrl-\_*).

All of the options mirror their counterparts in *jq*. The options are:

*-c*
	Use compact instead of pretty-printed output.

*-n*
	Don't read any input. Useful for using *ijq* as a calculator or to
	construct JSON data from scratch.

*-s*
	Read all input into a single array and apply the given filter to it.

*-r*
	If the filter output is a string it will be written directly to standard
	output rather than being formatted as a JSON string with quotes. Useful
	for using *ijq* in a pipeline with other programs that expect normal
	string input.

*-j*
	Implies *-r* and outputs without a trailing newline after each result.

*-a*
	Output strings by using only ASCII characters and escape sequences.

*-R*
	Don't parse the input as JSON, instead passing each line of input to the
	filter as a string. If combined with *-s* then the entire input is
	passed to the filter as a single long string.

*-M*
	Disable colored output.

*-C*
	Force colored output.

*-S*
	Output the fields of each object with the fields in sorted order.

*-f* _file_
	Read the filter from _file_. When this option is used, all positional
	arguments (if any) are interpreted as input files.

# CONFIG FILE

*ijq* reads configuration from _$XDG_CONFIG_HOME/ijq/config_. If
_$XDG_CONFIG_HOME_ is not set, then the fallback path is
_~/.config/ijq/config_.

The file format is *scfg*. All settings are optional. If a setting is present
both in the config file and on the command line, the command-line value takes
precedence.

*history-file* _file_
	Path to the history file. If set to '' then history is disabled.

*jq-bin* _file_
	Name of or path to the *jq* binary to use.

*hide-input-pane* _bool_
	If true, start with the input (left) viewing pane hidden.

*copy-filter-to-clipboard-on-exit* _bool_
	If true, copy the current filter to the clipboard when exiting with the
	*submit-filter* action (default: *Return*). The clipboard is only
	written after the output is written successfully to
	stdout. Exiting with *Ctrl-C* never writes to the clipboard.

*keymaps*
	Section containing key bindings. Any entry not set keeps its built-in
	default value.

	Bindings are written as key names (*Ctrl+N*, *Shift+Up*, *Return*) or
	single runes (*u*, *G*). _+_ and _-_ can both be used between modifiers
	and key names, for example *Ctrl+N* and *Ctrl-N*.

	When the filter input field has focus, printable runes are always treated
	as text input. To bind a global command while editing the filter, use a
	non-printable key or a modifier combination.

	Available entries are: *submit-filter*, *move-down*, *move-up*,
	*page-down*, *page-up*, *scroll-to-top*, *scroll-to-bottom*,
	*half-page-up*, *half-page-down*, *line-start*, *line-end*,
	*cursor-right*, *cursor-left*, *focus-input-pane*, *focus-output-pane*,
	*focus-filter-input*, *next-focus*, *previous-focus*,
	*toggle-input-pane*, *save-filter-history*, *next-autocomplete*,
	*previous-autocomplete*, *toggle-menu*, *copy-filter-to-clipboard*,
	*copy-output-to-clipboard*.

# KEY BINDINGS

These are the default key bindings. They can be overridden in the *keymaps*
section of the config file.

*Shift + Up*, *Shift + Left*
	Focus the input (left) viewing pane (*focus-input-pane*).

*Shift + Right*
	Focus the output (right) viewing pane (*focus-output-pane*).

*Shift + Down*
	Focus the text (filter) input field (*focus-filter-input*).

*Tab*
	When the text (filter) input field has focus, trigger autocompletion or
	navigate between the autocompletion list (*next-autocomplete*). When one
	of the viewing panes has focus, toggle between the different views
	(*next-focus*).

*Shift-Tab*
	Like *Tab*, but moves in the opposite direction
	(*previous-autocomplete*, *previous-focus*).

*Ctrl-A*
	When the text input field has focus, move the cursor to the beginning of
	the input. When one of the viewing panes has focus, set the column
	offset to 0 (*line-start*).

*Ctrl-E*
	When the text input field has focus, move the cursor to the end of the
	input. When one of the viewing panes has focus, set the column offset to
	view the longest visible line (*line-end*).

*Ctrl-F*
	When the text input field has focus, move the cursor one character
	forward (*cursor-right*). When one of the viewing panes has
	focus, scroll down one page (*page-down*).

*Ctrl-B*
	When the text input field has focus, move the cursor one character
	backward (*cursor-left*). When one of the viewing panes has
	focus, scroll up one page (*page-up*).

*Ctrl-D*
	When the text input field has focus, delete the character under the
	cursor. When one of the viewing panes has focus, scroll down one half
	page (*half-page-down*).

*Ctrl-U*
	When the text input field has focus, delete all of the text from the
	cursor to the beginning of the field. When one of the viewing panes has
	focus, scroll up one half page (*half-page-up*).

*Ctrl-O*
	Toggle visibility of the input (left) viewing pane
	(*toggle-input-pane*).

*Ctrl-S*
	Save the current filter to history and show a confirmation popup
	(*save-filter-history*).

*Ctrl-/*, *Ctrl-?*, *Ctrl-\_*
	Open or close the overlay menu (*toggle-menu*).

*Space*
	When the overlay root menu is open, activate the selected menu entry.
	When the Configure subview is open, toggle the selected option.

*Enter*, *Return*
	When the Configure subview is open, toggle the selected option.
	When the Manage history subview is open, apply the selected history
	entry to the filter and close the overlay (*submit-filter*).

*x*
	When the Manage history subview is open, delete the selected history
	entry after confirmation.

*/*
	When the Manage history subview is open, open a filter input for
	history entries.

*Esc*, *Ctrl-C*, *q*
	When overlay is open, go back to the root menu or close overlay.

*u*, *d*, *f*, *b*
	When one of the viewing panes has focus, scroll a half/full page
	up/down.

*g*
	When one of the viewing panes has focus, scroll to the top
	(*scroll-to-top*).

*G*
	When one of the viewing panes has focus, scroll to the bottom
	(*scroll-to-bottom*).

*Left*, *Down*, *Up*, *Right*
*h*, *j*, *k*, *l*
	When one of the viewing panes has focus, move the view
	left/down/up/right.

*Return*
	Close *ijq*. Write the contents of the output pane to stdout and the
	current input filter to stderr. The current input filter is also saved
	to the history file.

*Ctrl-C*
	When overlay is open, close overlay. When the filter input is focused
	and non-empty, clear the filter input. Otherwise, exit *ijq* immediately
	without writing anything to stdout or stderr.

# DEMO

See https://asciinema.org/a/496932 for a demo.

# SEE ALSO

jq(1)

# AUTHOR

Gregory Anders <greg@gpanders.com>
