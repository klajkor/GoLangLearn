// Package course embeds the curriculum (chapters, lessons and quizzes) so the
// server binary is self-contained.
package course

import "embed"

// FS holds one directory per chapter, e.g. "01-intro".
//
//go:embed all:[0-9][0-9]-*
var FS embed.FS
