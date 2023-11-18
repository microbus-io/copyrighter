# Copyrighter

`Copyrighter` is a utility that leverages Go's code generator to add a copyright notice to source files in the following languages:

* C (.c)
* C++ (.cpp)
* C# (.cs)
* CSS (.css)
* Go (.go)
* HTML (.html)
* Java (.java)
* JavaScript (.js)
* PHP (.php)
* PowerShell (.ps1)
* Python (.py)
* Shell (.sh)
* SQL (.sql)
* TypeScript (.ts)
* XML (.xml)
* YAML (.yaml)

# Usage

To install, `go get github.com/microbus-io/copyrighter`.

Create a `copyright.go` or `doc.go` file in each source code directory to be processed. Enter a comment with the copyright or license notice at the top of the file, followed by the `go generate` directive.

```go
/*
Copyright 2023 You
*/

//go:generate go run github.com/microbus-io/copyrighter

package yourpackage

import _ "github.com/microbus-io/copyrighter/i"
```

The first comment surrounded by `/*` and `*/` (on separate lines with nothing else added to those lines) or one where each line starts with `//` will be recognized as the copyright notice.

```go
/*
Good
*/

// Good
// Good

/* Bad */

var example /*
Bad
*/

var example // Bad
```

The following flags may be added to the `go:generate` directive:

* `-r` to recurse sub-directories
* `-v` for verbose output

Using the `-r` flag, it is enough to have a single `copyright.go` at the root of the project directory tree rather than in each directory. Nested `copyright.ignore` files can be used to exclude sub-directories.

# Copyright.ignore

A `copyright.ignore` file can be used to instruct the `Copyrighter` to ignore certain files. It supports the following patterns:

```sh
# Comment
file.ext
*.ext
*.*
*
subdir
subdir/file.ext
subdir/*.ext
subdir/*.*
subdir/*
```

Patterns are resolved to the directory where they are defined. For example, `*.go` excludes Go files only in the directory in which `copyright.ignore` is located and not in nested sub-directories.

The pattern `*` effectively prevents processing of the directory and all nested sub-directories.

# Legal

`Copyrighter` is released by `Microbus LLC` under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0).
