ijq
===

Interactive `jq` tool. Like [jqplay][] for the commandline.

[![asciicast](https://asciinema.org/a/a0pp8jmRNw74EIqTRS3pCvZUJ.svg)](https://asciinema.org/a/a0pp8jmRNw74EIqTRS3pCvZUJ)

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

Read from a file

    ijq file.json

Read from stdin

    curl -s https://api.github.com/users/gpanders | ijq

Press `Return` to close `ijq` and print the current filtered output to stdout.
You can use this in a pipe, .e.g.

    curl -s https://api.github.com/users/gpanders/repos | ijq > output.json

Press `Tab` or `Shift-Tab` to cycle through the windows. The display windows
can be scrolled using Vim-like bindings, i.e. `hjkl` or the arrow keys.

Similar Work
------------

- [jqplay][]
- [vim-jqplay][]

[vim-jqplay]: https://github.com/bfrg/vim-jqplay
