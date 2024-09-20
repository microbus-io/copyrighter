# Copyrighter

`Copyrighter` is a utility that leverages Go's code generator to add a copyright notice to source files. Because a legal notice today keeps the lawyers away.

It recognizes the following source file types:

* Bazel (.bazel)
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
* Terraform (.tf)
* TypeScript (.ts)
* XML (.xml)
* YAML (.yaml, *.yml)

# Usage

To install, `go get github.com/microbus-io/copyrighter`.

Place a `copyright.go` file in the root of the source code directory tree, with:

* A comment at the top of the file with the copyright or license notice
* A `go generate` directive
* File matching patterns (optional)
* `package` and `import` statements

Run `go generate` from the command line in the root directory of the project.

```go
/*
Copyright 2023 You
All rights reserved
*/

//go:generate go run github.com/microbus-io/copyrighter
// - *.*
// + *.go
// - /vendors/*

package yourpackage

import _ "github.com/microbus-io/copyrighter/i"
```

### Copyright Notice

The first comment surrounded by `/*` and `*/` (on separate lines with nothing else added to those lines) or one where each line starts with `//` will be recognized as the copyright notice.

```go
/*
Copyright 2023 You
All rights reserved
*/

...
```

or

```go
// Copyright 2023 You
// All rights reserved

...
```

The special constant `YYYY` may be used as placeholder for the current year.

```go
// Copyright 2023-YYYY You
```

### File Matching Patterns

By default, all recognized source file types are processed.
To customize which files to process, file matching patterns may be added anywhere in `copyright.go`. Patterns are executed in the order of their appearance: the last pattern that matches wins.

The following examples excludes all files by default, then re-includes `*.go` and `*.sql` files, except in the `/vendors` directory.

```go
// - *.*
// + *.go
// + *.sql
// - /vendors/*
```

A `-` or `+` is used to indicate if this pattern is an exclusion or inclusion pattern.

Patterns that start with a `/` are matched to the root directory where `copyright.go` is located. Otherwise, they are applied to any subdirectory.

The following pattern can be used to exclude hidden files on Unix:

```go
// - .*
```

The `Copyrighter` recurses into all descendant subdirectories, except those that contain their own `copyright.go` file with a `go:generate` directive.

### Verbose Flag

The `-v` flag may be added to the `go:generate` directive to produce verbose output.

```go
//go:generate go run github.com/microbus-io/copyrighter -v
```

# Legal

`Copyrighter` is released by `Microbus LLC` under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0).
