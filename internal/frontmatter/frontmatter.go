package frontmatter

import (
)

// Document is a note split into its frontmatter block and body. If hasFM = false,
// doc has no frontmatter.
type Document struct {
	hasFM bool
	newLine string
	openFence string
	preamble string
	entries []entry
	closeFence string
	body string // everything after the closing fence, unchanged
}

// entry is one top-level yaml key and the raw value
type entry struct {
	key string
	raw string
}

// Split parses src into a Document. It never errors. If the frontmatter is
// not well formatted, then it is considered body
func Split(src []byte) *Document {
}

// Patch merges set into the frontmatter, creating block if absent.
// Removes any keys named in unset. Existing keys are replaced in-place.
// New keys are always appended in sorted lexicographic order for determinism
//
// unset trumps set if both have the same key
func (d *Document) Patch(set map[string]any, unset []string) error {
}

func (d *Document) Render() []byte {
}

func (d *Fields() map[string]any {
}

func renderEntry(key string, val any, nl string) (string, error) {
}

func renderList(key string, items []any, nl string) (string, error) {
}

func scalarString(val any) (string, error) {
}

func needsQuote(s string) bool {
}

func looksNumeric(s string) bool {
}

func formatNumber(f float64) string {
}

// --- parsing ---

func parseValue(raw string) any {
}

func parseFlowList(inner string) []any {
}

func parseScalar(s string) any {
}

func topLevelKey(ln string) (string, bool) {
}

func splitLinesKeep(s string) []string {
}

func trimLineEnding(s string) string {
	return strings.TrimRight(s, "\r\n")
}

func detectNewLine(line string) string {
	if strings.HasSuffix(line, "\r\n") {
		return "\r\n"
	}
	return "\n"
}
