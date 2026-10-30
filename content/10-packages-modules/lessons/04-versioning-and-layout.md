---
title: Versions, Layout and Workspaces
minutes: 10
---
## Semantic versioning

Go modules use semantic versions written `vMAJOR.MINOR.PATCH`, such as `v1.4.2`.

| Part | Bump when... |
|---|---|
| MAJOR | you make incompatible API changes |
| MINOR | you add functionality in a backward-compatible way |
| PATCH | you make backward-compatible bug fixes |

Versions come from VCS tags (`git tag v1.4.2`). Versions below `v1.0.0` (for example `v0.9.1`) promise no stability. A dependency without tags is referenced by a *pseudo-version* like `v0.0.0-20240102150405-abcdef123456`.

## Major versions and import paths

Go has the **import compatibility rule**: a package with the same import path must stay backward compatible. So when you break the API, you must change the path. For major version 2 or higher, the module path ends in `/vN`:

```text
module github.com/you/lib/v2
```

```go
import "github.com/you/lib/v2/parser"
```

Versions `v0` and `v1` have **no** suffix. Because `lib` and `lib/v2` are different paths, a program can import both at once and migrate gradually.

> **Common mistake:** tagging `v2.0.0` without changing the module path to end in `/v2`. The go tool will refuse to treat it as a proper v2 module.

## Project layout

Go does not enforce a layout, but a few conventions are widely used:

```text
myapp/
  go.mod
  go.sum
  cmd/
    server/
      main.go          // package main: the server binary
    migrate/
      main.go          // package main: another binary
  internal/
    auth/
      auth.go
    store/
      store.go
  pkg/                  // optional: code meant for others to import
    client/
      client.go
  README.md
```

- **`cmd/<name>/`** holds one `main` package per executable. `go build ./cmd/server` builds just that one.
- **`internal/`** is enforced by the compiler: packages under an `internal` directory can only be imported by code rooted at the parent of `internal`. Outsiders cannot import `myapp/internal/auth`, so you are free to change it.
- **`pkg/`** is a convention, not a rule. Many projects skip it.

For a small project, a flat layout with a `main.go` and a few files is perfectly fine. Start simple and split packages when there is a real reason, such as a clear responsibility or a reusable library.

> **Tip:** name packages after what they provide (`store`, `auth`), not after a category (`utils`, `common`, `helpers`). Those catch-all names make code harder to understand.

## Vendoring

`go mod vendor` copies all dependencies into a `vendor/` directory inside your repo. Builds then use that directory by default when `vendor/` exists and the `go` version in `go.mod` is 1.14 or newer. It is useful for offline builds or strict supply-chain review, at the cost of a bigger repository.

```bash
go mod vendor
go build -mod=vendor ./...
```

## Workspaces

When you work on several modules at once (say an app and a library it depends on), you can avoid temporary `replace` lines with a **workspace**, introduced in Go 1.18:

```bash
go work init ./app ./lib
```

This creates `go.work`:

```text
go 1.22

use (
    ./app
    ./lib
)
```

While `go.work` is present, the go command resolves those modules from your local directories. Typically you do **not** commit `go.work`; it is a local development convenience. Use `go work sync` to sync dependency versions to the workspace modules and `GOWORK=off` to ignore it.

## Build constraints (brief)

Files can be included conditionally by a build tag:

```go
//go:build linux && amd64

package platform
```

A filename like `store_windows.go` or `store_linux.go` is also picked automatically by `GOOS`. Files ending in `_test.go` are compiled only for tests.

## Key takeaways

- Versions follow semver and come from VCS tags; `v0` has no stability promise.
- For v2 and above, the module path and import paths end in `/v2`, `/v3`, and so on.
- `cmd/` holds binaries; `internal/` is compiler-enforced privacy; `pkg/` is optional convention.
- Avoid package names such as `utils` and `common`.
- `go mod vendor` copies dependencies into your repo for offline or audited builds.
- `go.work` workspaces let several local modules develop together without `replace`.
