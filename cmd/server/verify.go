package main

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

var assetRef = regexp.MustCompile(`[A-Za-z0-9_.-]+\.(?:js|css)`)

// verifyFrontend checks that every script, stylesheet and lazily loaded chunk
// that the built UI refers to is actually present in fsys. A half-copied or
// stale web/dist would otherwise show up as a blank page with no explanation.
func verifyFrontend(fsys fs.FS) error {
	index, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		return fmt.Errorf("index.html is missing")
	}
	var missing []string
	seen := map[string]bool{}
	var visit func(name string, body []byte)
	visit = func(name string, body []byte) {
		for _, ref := range assetRef.FindAll(body, -1) {
			file := "assets/" + string(ref)
			if seen[file] || !strings.Contains(string(ref), "-") {
				continue // not a fingerprinted build file
			}
			seen[file] = true
			data, err := fs.ReadFile(fsys, file)
			if err != nil {
				missing = append(missing, file)
				continue
			}
			if strings.HasSuffix(file, ".js") {
				visit(file, data)
			}
		}
	}
	visit("index.html", index)
	if len(missing) > 0 {
		return fmt.Errorf("the web UI build is incomplete (missing %s)", strings.Join(missing, ", "))
	}
	return nil
}
