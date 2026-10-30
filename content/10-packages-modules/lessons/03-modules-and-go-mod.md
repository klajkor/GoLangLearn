---
title: Modules, go.mod and go.sum
minutes: 10
---
A **module** is a collection of packages that are versioned and released together. It is defined by a `go.mod` file at its root. Since Go 1.16 modules are the default way to build Go code, so every project starts with one.

## Creating a module

```bash
mkdir hello && cd hello
go mod init example.com/hello
```

This creates `go.mod`:

```text
module example.com/hello

go 1.22
```

The **module path** (`example.com/hello`) becomes the prefix of every import path inside the module. If you publish the code, the path usually matches the repository URL, for example `github.com/you/project`. For private experiments, any name works.

A package at `hello/greet` is therefore imported as `example.com/hello/greet`.

## What go.mod contains

```text
module github.com/you/shop

go 1.22

require (
    github.com/google/uuid v1.6.0
    golang.org/x/text v0.14.0 // indirect
)

replace example.com/old => ../old
exclude example.com/bad v1.2.3
```

| Directive | Meaning |
|---|---|
| `module` | The module's own path |
| `go` | Minimum Go language version for this module |
| `toolchain` | Optional preferred toolchain version |
| `require` | A dependency and the minimum version needed |
| `replace` | Substitute a module with another version or a local directory |
| `exclude` | Forbid a specific version |

An `// indirect` comment marks a dependency your code does not import directly; it is needed by one of your dependencies.

## go.sum

`go.sum` lists cryptographic hashes of every module version your build uses. Go checks downloaded code against these hashes, so a tampered or changed release is detected. Commit **both** `go.mod` and `go.sum`. You never edit `go.sum` by hand.

## The daily commands

```bash
go get github.com/google/uuid@v1.6.0   # add or change a dependency
go get -u ./...                        # upgrade dependencies to newer minor/patch
go mod tidy                            # sync go.mod/go.sum with your imports
go mod download                        # fetch modules into the local cache
go mod why github.com/google/uuid      # why is this needed?
go list -m all                         # list every module in the build
```

`go mod tidy` is the one to remember: it **adds** requirements for imports that are missing and **removes** ones nobody uses. Run it before committing.

## Minimal version selection

If your module needs `A v1.2.0` and another dependency needs `A v1.4.0`, Go picks **v1.4.0**: the highest of the *minimum* versions required by anyone. This is called minimal version selection (MVS). It is deterministic: no lock file is needed to get reproducible builds, and nothing upgrades just because a newer release appeared. `go.mod` plus `go.sum` fully describe the build.

## Where modules live

Downloaded modules are cached read-only under `$GOPATH/pkg/mod` (use `go env GOMODCACHE` to see it). By default `go get` fetches through the module proxy (`proxy.golang.org`) and verifies hashes against a public checksum database. For private repositories, set `GOPRIVATE`:

```bash
go env -w GOPRIVATE=github.com/mycompany/*
```

## A tiny two-package module

```text
hello/
  go.mod                  module example.com/hello
  main.go                 package main
  greet/
    greet.go              package greet
```

```go
// main.go
package main

import (
    "fmt"

    "example.com/hello/greet"
)

func main() {
    fmt.Println(greet.Hello("Gopher"))
}
```

```go
// greet/greet.go
package greet

func Hello(name string) string { return "Hello, " + name + "!" }
```

```bash
go run .
go build -o hello .
```

```text
Hello, Gopher!
```

> **Common mistake:** running `go run main.go` in a package that has several files. Use `go run .` so the whole package is compiled.

## Key takeaways

- A module is defined by `go.mod`; its module path prefixes all import paths in it.
- Create one with `go mod init <path>`.
- `go.sum` holds verification hashes; commit it, never hand-edit it.
- `go get` changes dependencies; `go mod tidy` cleans up `go.mod` and `go.sum`.
- Minimal version selection picks the highest required minimum, so builds are reproducible without a lock file.
- Use `go run .` and `go build` on packages, not single files, in real projects.
