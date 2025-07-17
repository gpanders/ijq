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

*ijq* maintains a history of used filters, unless disabled with the *-H* option.
Delete all text in the filter field to browse any available history.

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

*-H* _file_
	Specify the path to store history. If set to '' (-H ''), then history
	will not be captured.

# KEY BINDINGS

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
