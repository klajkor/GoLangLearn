---
title: "Putting It Together: Text Processing"
minutes: 7
---
Let's combine runes, `strings`, `strconv`, maps and `fmt` into small, realistic text tasks.

## Word frequency

Count how often each word appears, then print the result in a stable order. Maps have random iteration order, so we sort the keys before printing.

```go run
package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	text := "the cat and the hat and THE bat"

	counts := map[string]int{}
	for _, w := range strings.Fields(strings.ToLower(text)) {
		counts[w]++
	}

	words := make([]string, 0, len(counts))
	for w := range counts {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if counts[words[i]] != counts[words[j]] {
			return counts[words[i]] > counts[words[j]]
		}
		return words[i] < words[j]
	})

	for _, w := range words {
		fmt.Printf("%-5s %d\n", w, counts[w])
	}
}
```

We sort by count descending and break ties alphabetically, so the output never changes between runs.

## Parsing key=value lines

`Cut`, `TrimSpace` and `Atoi` make a tiny config parser. Note how errors are handled at each step.

```go run
package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parse(line string) (string, int, error) {
	key, val, ok := strings.Cut(line, "=")
	if !ok {
		return "", 0, fmt.Errorf("no '=' in %q", line)
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return "", 0, fmt.Errorf("bad value for %s: %w", strings.TrimSpace(key), err)
	}
	return strings.TrimSpace(key), n, nil
}

func main() {
	for _, line := range []string{"port = 8080", "retries=3", "oops", "limit = ten"} {
		k, v, err := parse(line)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Printf("%s -> %d\n", k, v)
	}
}
```

## Working with runes: palindromes

Compare characters, not bytes, and ignore case and punctuation.

```go run
package main

import (
	"fmt"
	"unicode"
)

func isPalindrome(s string) bool {
	var rs []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		if rs[i] != rs[j] {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println(isPalindrome("A man, a plan, a canal: Panama"))
	fmt.Println(isPalindrome("Gopher"))
	fmt.Println(isPalindrome("été"))
}
```

The `unicode` package classifies runes (`IsLetter`, `IsDigit`, `IsSpace`, `IsUpper`) and converts case (`ToUpper`, `ToLower`).

> **Tip:** `bytes` has a twin of almost every `strings` function (`bytes.Contains`, `bytes.Split`, `bytes.Buffer`). Use it when your data is already `[]byte`, such as network or file content.

## A checklist for string code

- Need characters? Use `range` or `[]rune`, not byte indexing.
- Building in a loop? Use `strings.Builder`.
- Parsing input? Check every `error` from `strconv`.
- Printing a map or a report? Sort first, then use width flags to align.
- Debugging odd text? Print it with `%q` to reveal whitespace and escapes.

## Key takeaways

- Combine `Fields`, `ToLower` and a map to count words; sort keys for deterministic output.
- `Cut` plus `TrimSpace` plus `Atoi` is a compact `key=value` parser.
- `unicode` helps process runes correctly across languages.
- `%q` is the best way to inspect a string while debugging.
