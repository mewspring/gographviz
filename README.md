## Fork

This is a fork of the [gographviz] package.

## WIP

This project is a *work in progress*. The implementation is *incomplete* and subject to change. The documentation may be inaccurate.

# dot

[![GoDoc](https://godoc.org/github.com/mewspring/dot?status.svg)](https://godoc.org/github.com/mewspring/dot)

The dot project implements a graphviz DOT language parser, which was generated using the [gocc] compiler toolkit.

## Changes

* Add dominator tree support (based on a modified version of [x/tools/go/ssa][ssa].
* Return [*Graph](https://godoc.org/github.com/mewspring/dot#Graph) instead of [Interface](https://godoc.org/github.com/mewspring/dot#Interface) from the [Read](https://godoc.org/github.com/mewspring/dot#Read) and [NewAnalysedGraph](https://godoc.org/github.com/mewspring/dot#NewAnalysedGraph) functions.

[x/tools/go/ssa]: https://godoc.org/golang.org/x/tools/go/ssa

## Public domain

Any changes (starting from rev 9ad29961d022b112cf0608e5e7012a3ede1b0f7f) made to the [gographviz] repository are hereby released into the [public domain], including changes to the source code and additions of any original content.

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/

## License

The original [gographviz] source code is goverend by an [Apache license](LICENSE).

Portions of [gocc]'s source code have been derived from Go, and are covered by a [BSD license](http://golang.org/LICENSE).

The dominator tree support is directly derived from [x/tools/go/ssa], which is part of the Go project and governed by a [BSD license](http://golang.org/LICENSE).

[gographviz]: https://github.com/awalterschulze/gographviz
[gocc]: https://github.com/goccmack/gocc
