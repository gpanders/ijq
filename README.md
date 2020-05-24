ijq
===

Interactive `jq` tool. Like [jqplay][] for the commandline.

[![asciicast](https://asciinema.org/a/333292.svg)](https://asciinema.org/a/333292)

[jqplay]: https://jqplay.org

Building
--------

Install [go][golang]. Then simply run

    make

from the root of the project.

[golang]: https://golang.org/dl/

Installation
------------

To install `ijq` under `/usr/local/bin/` simply use

    make install

To install to another location, set the `PREFIX` variable, e.g.

    make PREFIX=~/.local install

Usage
-----

ijq uses [jq][] under the hood, so make sure you have that installed first.

Read from a file

    ijq file.json

Read from stdin

    curl -s https://api.github.com/users/gpanders | ijq

Press `Return` to close `ijq` and print the current filtered output to stdout.
You can use this in a pipe, .e.g.

    curl -s https://api.github.com/users/gpanders/repos | ijq > output.json

This will also print the current filter to stderr. This allows you to save the
filter for re-use with `jq` in the future:

    ijq file.json 2>filter.txt

    # Same output as above
    jq -f filter.txt file.json

Press `Tab` or `Shift-Tab` to cycle through the windows. The display windows
can be scrolled using Vim-like bindings, i.e. `hjkl` or the arrow keys.

[jq]: https://stedolan.github.io/jq/

Similar Work
------------

- [jqplay][]
- [vim-jqplay][]

[vim-jqplay]: https://github.com/bfrg/vim-jqplay
