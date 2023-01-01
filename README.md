# Copyrighter

`Copyrighter` is a utility that adds a copyright notice to source files using Go's code-generator. By default, it adds the notice to the following languages (file extensions):

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
* TypeScript (.ts)
* XML (.xml)
* YAML (.yaml)

To install, `go get github.com/microbus-io/copyrighter`.

Create a `copyright.go` file in each source code directory to be processed.
Enter a comment with the copyright or license notice followed by the `go generate` directive.

```go
/*
Copyright 2022 You
*/

//go:generate go run github.com/microbus-io/copyrighter

package yourpackage
```

The following flags may be added to the directive:

* `-r` to recurse sub-directories
* `-v` for verbose output
* `-x yaml,html,etc` to exclude files by extension

`Copyrighter` is released under the Apache 2 license.
