ijq
===

Interactive `jq` tool. Like [jqplay] for the commandline.

[![asciicast](https://asciinema.org/a/333292.svg)](https://asciinema.org/a/333292)

[jqplay]: https://jqplay.org

Installation
------------

### Homebrew

If you use macOS and Homebrew, you can install `ijq` with

    brew install gpanders/tap/ijq

### Download a release

Select the version you want to download from [sourcehut] and download one of
the precompiled releases from that page. Then extract the archive and copy the
binary and, optionally, the man page to the correct location.

Example:

    wget https://git.sr.ht/~gpanders/ijq/refs/vX.Y.Z/ijq-vX.Y.Z-linux-x86_64.tar.gz
    tar xf ijq-vX.Y.Z-linux-x86_64.tar.gz
    cd ijq-vX.Y.Z
    cp ijq /usr/local/bin
    mkdir -p /usr/local/share/man/man1
    cp ijq.1 /usr/local/share/man/man1

[sourcehut]: https://git.sr.ht/~gpanders/ijq/refs

### Build from source

Install [go]. To install `ijq` under `/usr/local/bin/` simply run

    make install

from the root of the project. To install to another location, set the `prefix`
variable, e.g.

    make prefix=~/.local install

To generate the man page you will also need to have [scdoc] installed.

[go]: https://golang.org/dl/
[scdoc]: https://sr.ht/~sircmpwn/scdoc

Usage
-----

ijq uses [jq] under the hood, so make sure you have that installed first.

Read from a file:

    ijq file.json

Read from stdin:

    curl -s https://api.github.com/users/gpanders | ijq

Press `Return` to close `ijq` and print the current filtered output to stdout.
This will also print the current filter to stderr. This allows you to save the
filter for re-use with `jq` in the future:

    ijq file.json 2>filter.jq

    # Same output as above
    jq -f filter.jq file.json

Press `Tab` or `Shift-Tab` to cycle through the windows. The display windows
can be scrolled using Vim-like bindings, i.e. `hjkl` or the arrow keys.

You can configure the colors by setting the `JQ_COLORS` environment variable.
See the [jq documentation][colors] for more details.

[jq]: https://stedolan.github.io/jq/
[colors]: https://stedolan.github.io/jq/manual/#Colors

Similar Work
------------

- [jqplay]
- [vim-jqplay]

[vim-jqplay]: https://github.com/bfrg/vim-jqplay
