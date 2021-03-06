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

# NAVIGATION

Use Shift + arrow keys to navigate between the different panes. The text panes
can be scrolled using the arrow keys or vi-style bindings (e.g. j/k to scroll
up/down by line, Ctrl-F/Ctrl-B to scroll up/down by half page).

Delete all text in the filter field to browse any available history.

Use Ctrl-C to exit *ijq* immediately, discarding all state.

# DEMO

See https://asciinema.org/a/bIyt8DWu3L3QAP1BMKefoaUob for a demo.

# SEE ALSO

jq(1)

# BUGS AND CONTRIBUTING

Report bugs to https://todo.sr.ht/~gpanders/ijq (or send an email to
~gpanders/ijq@todo.sr.ht).

Patches can be sent to ~gpanders/ijq-devel@lists.sr.ht.

# AUTHOR

Gregory Anders <greg@gpanders.com>
