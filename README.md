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
* `-x yaml,html,etc` to exclude files by extension

`Copyrighter` is released under the [Apache 2.0 license](http://www.apache.org/licenses/LICENSE-2.0).
