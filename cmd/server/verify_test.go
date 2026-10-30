package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestVerifyFrontend(t *testing.T) {
	good := fstest.MapFS{
		"index.html":            {Data: []byte(`<script src="/assets/index-AAAA.js"></script><link href="/assets/index-BBBB.css">`)},
		"assets/index-AAAA.js":  {Data: []byte(`import("./Chunk-CCCC.js")`)},
		"assets/index-BBBB.css": {Data: []byte("body{}")},
		"assets/Chunk-CCCC.js":  {Data: []byte("export default 1")},
	}
	if err := verifyFrontend(good); err != nil {
		t.Errorf("complete build rejected: %v", err)
	}

	delete(good, "assets/Chunk-CCCC.js")
	err := verifyFrontend(good)
	if err == nil || !strings.Contains(err.Error(), "assets/Chunk-CCCC.js") {
		t.Errorf("missing lazy chunk not reported: %v", err)
	}

	if err := verifyFrontend(fstest.MapFS{}); err == nil {
		t.Error("empty dist accepted")
	}
}
