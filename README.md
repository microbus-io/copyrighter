# Copyrighter

`Copyrighter` is a utility that adds a copyright notice to source files using Go's code generator. Notices are added to the following languages (file extensions):

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
```

To be recognized, the comment must be surrounded by `/*` and `*/` on separate lines with nothing else added to those lines; or alternatively, each line of the comment must start with `//`.

```go
/*
Good
*/

// Good
// Good

/* Bad */

var x /*
Bad
*/

var x // Bad
```

The following flags may be added to the `go:generate` directive:

* `-r` to recurse sub-directories
* `-v` for verbose output
* `-x yaml,html,etc` to exclude files by extension

`Copyrighter` is released under the Apache 2 license.
