// Package skills embeds the hush SKILL.md for offline access.
package skills

import _ "embed"

//go:embed SKILL.md
var Content []byte
