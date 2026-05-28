# Copyrighter

`Copyrighter` is a Go build tool that adds a copyright notice to source files. Because a legal notice today keeps the lawyers away.

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

# Installation

`Copyrighter` is distributed as a Go build tool, which requires Go 1.24 or later.
From the root directory of your project, add it as a tool dependency:

```
go get -tool github.com/microbus-io/copyrighter
```

This records the tool in your `go.mod` so the version is pinned alongside your other dependencies.

# Usage

Place a `COPYRIGHT` file in the root of the source code directory tree, with:

* The copyright or license notice at the top of the file
* A `---` divider
* File matching patterns (optional)

```
Copyright 2023-yyyy You
All rights reserved

---

!*.*
*.go
!/vendors/*
```

Then run the tool from the root directory of your project:

```
go tool copyrighter
```

### Copyright Notice

The notice is the literal text above the `---` divider. Blank lines within the notice are preserved; trailing blank lines are trimmed.

The special constant `YYYY` may be used as a placeholder for the current year, and `yyyy` for the year in which the file was last modified.

```
Copyright 2023-YYYY You
All rights reserved
```

### File Matching Patterns

By default, all recognized source file types are processed.
To customize which files to process, file matching patterns may be added after the `---` divider in `COPYRIGHT`. Patterns are evaluated in order: the last pattern that matches wins.

Patterns use a `.gitignore`-style syntax, but with the include/exclude semantics inverted:

* A bare pattern **includes** matching files.
* A pattern prefixed with `!` **excludes** matching files.
* Lines beginning with `#` and blank lines are ignored.

The following example excludes all files by default, then re-includes `*.go` and `*.sql` files, except in the `/vendors` directory:

```
!*.*
*.go
*.sql
!/vendors/*
```

Patterns that start with a `/` are anchored to the root directory where `COPYRIGHT` is located. Otherwise, they apply at any depth.

The following pattern excludes hidden files on Unix:

```
!.*
```

The `Copyrighter` recurses into all descendant subdirectories, except those that contain their own `COPYRIGHT` file.

### Verbose Flag

Pass the `-v` flag to produce verbose output:

```
go tool copyrighter -v
```

# Legal

`Copyrighter` is released by `Microbus LLC` under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0).
