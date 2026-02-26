ijq(1)

# NAME

ijq - interactive jq

# SYNOPSIS

*ijq* [*-cnsrRMSV*] [*-f* _file_] [_filter_] [_files ..._]

# DESCRIPTION

*ijq* is a near drop-in replacement for *jq* that allows you to interactively
see the results of your filter as you construct it.

*ijq* contains two panes and an input field: the left pane is the original,
unmodified input data and the right pane contains the filtered output. When you
are finished, press Return or Ctrl-C to exit. The filtered output will be
written to standard output and the filter itself will be written to standard
error.

*ijq* maintains a history of used filters unless disabled via
*history-file* in the config file. Delete all text in the filter field to
browse any available history.

If _files_ is omitted then *ijq* reads data from standard input.

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

*-R*
	Don't parse the input as JSON, instead passing each line of input to the
	filter as a string. If combined with *-s* then the entire input is
	passed to the filter as a single long string.

*-M*
	Disable colored output.

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
	*page-down*, *line-start*, *line-end*, *half-page-up*,
	*half-page-down*, *filter-cursor-right*, *filter-cursor-left*,
	*focus-input-pane-up*, *focus-input-pane-left*, *focus-output-pane*,
	*focus-filter-input*, *next-focus*, *previous-focus*,
	*toggle-input-pane*, *textview-page-up*, *textview-page-up-alt*,
	*textview-page-down*, and *textview-end*.

# KEY BINDINGS

These are the default key bindings. They can be overridden in the *keymaps*
section of the config file.

*Shift + Up*, *Shift + Left*
	Focus the input (left) viewing pane.

*Shift + Right*
	Focus the output (right) viewing pane.

*Shift + Down*
	Focus the text input field.

*Tab*
	When the text input field has focus, navigate between the autocompletion
	list. When one of the viewing panes has focus, toggle between the
	different views.

*Shift-Tab*
	Like *Tab*, but moves in the opposite direction.

*Ctrl-A*
	When the text input field has focus, move the cursor to the beginning of
	the input. When one of the viewing panes has focus, set the column
	offset to 0.

*Ctrl-E*
	When the text input field has focus, move the cursor to the end of the
	input. When one of the viewing panes has focus, set the column offset to
	view the longest visible line.

*Ctrl-F*
	When the text input field has focus, move the cursor one character
	forward. When one of the viewing panes has focus, scroll down one page.

*Ctrl-B*
	When the text input field has focus, move the cursor one character
	backward. When one of the viewing panes has focus, scroll up one page.

*Ctrl-D*
	When the text input field has focus, delete the character under the
	cursor. When one of the viewing panes has focus, scroll down one half
	page.

*Ctrl-U*
	When the text input field has focus, delete all of the text from the
	cursor to the beginning of the field. When one of the viewing panes has
	focus, scroll up one half page.

*Ctrl-O*
	Toggle visibility of the input (left) viewing pane.

*u*, *d*, *f*, *b*
	When one of the viewing panes has focus, scroll a half/full page
	up/down.

*Left*, *Down*, *Up*, *Right*
*h*, *j*, *k*, *l*
	When one of the viewing panes has focus, move the view
	left/down/up/right.

*Return*
	Close *ijq*. Write the contents of the output pane to stdout and the
	current input filter to stderr. The current input filter is also saved
	to the history file.

*Ctrl-C*
	Exit *ijq* immediately, discarding all state.

# DEMO

See https://asciinema.org/a/496932 for a demo.

# SEE ALSO

jq(1)

# AUTHOR

Gregory Anders <greg@gpanders.com>
