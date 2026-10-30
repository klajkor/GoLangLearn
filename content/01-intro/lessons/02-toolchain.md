---
title: Installing Go and the Toolchain
minutes: 8
---
Go ships as a single download that includes the compiler, the standard library, and a set of tools all driven by one command: `go`. There is nothing else to install to get started.

## Installing

Download the installer for your platform from [go.dev/dl](https://go.dev/dl/), or use a package manager (`brew install go`, `winget install GoLang.Go`, `apt install golang`). Then check it works in a new terminal:

```bash
go version
```

```text
go version go1.22.5 windows/amd64
```

Your version number will differ. This course assumes Go 1.22 or newer.

## Modules instead of GOPATH

Old tutorials talk about `GOPATH`, a single workspace where all code had to live. You can ignore that: modern Go uses **modules**. A module is a folder tree with a `go.mod` file at its root, and it can live anywhere on disk.

Create a new project like this:

```bash
mkdir hello
cd hello
go mod init example.com/hello
```

That writes a tiny `go.mod`:

```text
module example.com/hello

go 1.22
```

- `module` is the module path: the name other code uses to import your packages. For code you will publish it is usually the repository URL such as `github.com/you/project`; for experiments any name works.
- `go` states the language version the module expects.
- Later, `require` lines appear here when you add third-party dependencies (`go get`), and a `go.sum` file records checksums so downloads are verified.

## The everyday commands

| Command | What it does |
|---|---|
| `go run .` | Compiles to a temporary binary and runs it. Perfect while learning. |
| `go build` | Compiles and writes an executable into the current folder. |
| `go install` | Builds and puts the binary in your `GOBIN` (default `~/go/bin`). |
| `go fmt ./...` | Rewrites source files into the one official layout. |
| `go vet ./...` | Reports suspicious code that compiles but is probably wrong. |
| `go test ./...` | Runs tests (a later chapter). |
| `go mod tidy` | Adds missing and removes unused dependencies from `go.mod`. |

The `./...` pattern means "this folder and everything below it".

## go fmt: the end of style debates

Go has exactly one formatting style, applied by `gofmt`. Tabs for indentation, spacing around operators, sorted imports: the tool decides. Most editors run it on save. You never argue about formatting in code review.

## go vet: a cheap safety net

The compiler rejects errors, but some mistakes are legal Go and still wrong. For example, `fmt.Printf("%d", "text")` compiles, yet prints garbage. `go vet` catches this class of bug:

```go
fmt.Printf("%d items\n", "three")
```

```text
./main.go:8:2: fmt.Printf format %d has arg "three" of wrong type string
```

> **Tip:** `go build` and `go run` do not run `go vet` for you, but `go test` runs a subset of its checks. Make `go vet ./...` part of your habit and your CI.

## Cross-compiling

Because Go compiles to native code without a runtime to install, you can build for another platform just by setting environment variables:

```bash
GOOS=linux GOARCH=arm64 go build -o app-linux-arm64
```

The result is a single file you can copy to a server and run.

> **Common mistake:** running `go run main.go` in a project with several files. That compiles only the file you named. Use `go run .` to build the whole package in the current folder.

## Key takeaways

- One `go` command covers running, building, formatting, vetting, testing, and dependency management.
- Projects are modules defined by `go.mod` (`go mod init <path>`); no GOPATH setup is needed.
- `go run .` for quick runs, `go build` for a binary, `go fmt` and `go vet` to keep code clean.
- Go cross-compiles with `GOOS` and `GOARCH`, producing a single static executable.
