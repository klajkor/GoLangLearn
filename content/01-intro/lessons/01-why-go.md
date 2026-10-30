---
title: Why Go?
minutes: 6
---
Go (often called Golang) was designed at Google starting in 2007 by Robert Griesemer, Rob Pike and Ken Thompson, and released as open source in 2009. They were tired of waiting for huge C++ builds, wrestling with tangled dependencies and fighting complexity in large codebases. Go is their answer: a small language that compiles quickly, runs fast, and stays readable after years of teamwork.

## What makes Go different

| Property | What it means for you |
|---|---|
| Compiled, statically typed | Errors are caught before the program runs; you ship one native binary. |
| Fast compilation | Even big projects build in seconds, so the edit-run loop feels like a scripting language. |
| Garbage collected | No manual `free`; memory safety without the ceremony of C. |
| Built-in concurrency | Goroutines and channels are part of the language, not a library. |
| Small language | About 25 keywords. There is usually one obvious way to do something. |
| Batteries included | HTTP servers, JSON, testing, and formatting ship in the standard library and toolchain. |

Go deliberately leaves things out: there is no class inheritance, no exceptions, and no operator overloading. Errors are ordinary values, and composition replaces inheritance. This feels restrictive for a day or two and then feels liberating, because every Go codebase looks familiar.

## Where Go shines

Go is a favourite for network services and APIs, command-line tools, cloud infrastructure (Docker, Kubernetes and Terraform are written in Go), and anything that needs to handle many things at once. It is a weaker fit for GUI-heavy desktop apps or heavy numeric work where Python's or C++'s ecosystems are richer.

## A first taste of concurrency

You do not need to understand this yet, but it shows why people like Go. Each `go` statement starts a lightweight goroutine, and a `sync.WaitGroup` waits for them all to finish. Every goroutine writes to its own slot, so the result is the same on every run.

```go run
package main

import (
	"fmt"
	"sync"
)

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	squares := make([]int, len(numbers))

	var wg sync.WaitGroup
	for i, n := range numbers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			squares[i] = n * n
		}()
	}
	wg.Wait()

	fmt.Println(squares)
}
```

Five goroutines computed the squares at the same time. We will study goroutines properly in a later chapter.

> **Tip:** Go is also famous for its backward-compatibility promise. Code written for Go 1.0 in 2012 still compiles today, so what you learn here does not expire quickly.

## How this course works

Each chapter has short lessons with runnable examples. Press **Run** on any editable block, change the code and run it again; breaking things on purpose is the fastest way to learn. A quiz at the end of each chapter checks that the ideas stuck.

## Key takeaways

- Go is a compiled, statically typed, garbage-collected language built for simplicity and fast builds.
- It trades features (inheritance, exceptions) for readability and consistency.
- Concurrency is built in through goroutines and channels.
- It produces a single self-contained binary, which makes deployment easy.
- Go is especially strong for services, tools, and cloud infrastructure.
