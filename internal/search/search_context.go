package search

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

// SearchContext extracts context around the best match for query terms in content.
// Returns up to windowRunes characters centered on the match, with "..."
// prefixed/suffixed if truncated. Returns empty string if no terms match
func SearchContext(content, query string, windowRunes int) string {
	terms := tokenize(query)
	if len(terms) == 0 {
		return ""
	}

	contentLower := strings.ToLower(content)

	// Find the best match position (first occurence of any term)
	bestPos := -1
	bestLen := 0
	for _, term := range terms {
		pos := strings.Index(contentLower, term)
		if pos != -1 && (bestPos == -1 || pos < bestPos) {
			bestPos = pos
			bestLen = len(term)
		}
	}

	if bestPos == -1 {
		return truncateRunes(content, windowRunes)
	}

	// Return window around best match
	return extractWindow(content, bestPos, bestLen, windowRunes)
}

func clamp(i, min, max int) int {
	if i <= max && i >= min {
		return i
	}
	if i > max {
		return max
	}
	return min
}

func extractWindow(content string, pos, matchLen, windowRunes int) string {
	// Convert byte positions to rune positions for proper unicode handling
	runePos := utf8.RuneCountInString(content[:pos])
	totalRunes := utf8.RuneCountInString(content)

	// Calculate start and end rune pos
	halfWindow := windowRunes / 2
	startRune := clamp(runePos - halfWindow, 0, totalRunes)
	endRune := clamp(startRune + windowRunes, 0, totalRunes)
	startRune = clamp(endRune - windowRunes, 0, totalRunes)

	start := runeOffsetToByteOffset(content, startRune)
	end := runeOffsetToByteOffset(content, endRune)
	start = snapToWordStart(content, start)
	end = snapToWordEnd(content, end)

	searchContext := content[start:end]
	searchContext = strings.TrimSpace(searchContext)

	var b strings.Builder
	if start > 0 {
		b.WriteString("...")
	}
	b.WriteString(searchContext)
	if end < len(content) {
		b.WriteString("...")
	}
	return b.String()
}

func runeOffsetToByteOffset(s string, runeOffset int) int {
	byteOffset := 0
	for i := 0; i < runeOffset && byteOffset < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[byteOffset:])
		byteOffset += size
	}
	return byteOffset
}

func snapToWordStart(s string, pos int) int {
	// Move backwards to find word boundary
	for pos > 0 {
		r, size := utf8.DecodeLastRuneInString(s[:pos])
		if unicode.IsSpace(r) {
			break
		}
		pos -= size
	}
	return pos
}

func snapToWordEnd(s string, pos int) int {
	// Move forwards to word boundary
	for pos < len(s) {
		r, size := utf8.DecodeRuneInString(s[pos:])
		if unicode.IsSpace(r) {
			break
		}
		pos += size
	}
	return pos
}

func truncateRunes(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return strings.TrimSpace(s)
	}
	end := runeOffsetToByteOffset(s, maxRunes)
	end = snapToWordEnd(s, end)
	return strings.TrimSpace(s[:end]) + "..."
}
