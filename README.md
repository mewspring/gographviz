## Fork

This is a fork of the [gographviz] package.

# dot

[![GoDoc](https://godoc.org/github.com/mewfork/dot?status.svg)](https://godoc.org/github.com/mewfork/dot)

The dot project implements a graphviz DOT language parser, which was generated using the [gocc] compiler toolkit.

## Changes

* Add dominator tree support (based on a modified version of [x/tools/go/ssa][ssa].
* Return [*Graph](https://godoc.org/github.com/mewfork/dot#Graph) instead of [Interface](https://godoc.org/github.com/mewfork/dot#Interface) from the [Read](https://godoc.org/github.com/mewfork/dot#Read) and [NewAnalysedGraph](https://godoc.org/github.com/mewfork/dot#NewAnalysedGraph) functions.

[x/tools/go/ssa]: https://godoc.org/golang.org/x/tools/go/ssa

## Public domain

Any changes (starting from rev 9ad29961d022b112cf0608e5e7012a3ede1b0f7f) made to the [gographviz] repository are hereby released into the [public domain], including changes to the source code and additions of any original content.

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/

## License

The original [gographviz] source code is goverend by an [Apache license](LICENSE).

Portions of [gocc]'s source code have been derived from Go, and are covered by a [BSD license](http://golang.org/LICENSE).

The dominator tree support is directly derived from [x/tools/go/ssa], which is part of the Go project and governed by a [BSD license](http://golang.org/LICENSE).

[gographviz]: https://code.google.com/p/gographviz/
[gocc]: https://code.google.com/p/gocc/
