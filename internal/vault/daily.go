package vault

import (
	"strings"
	"time"
)

type DailyNoteConfig struct {
	Folder   string `json:"folder"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

const dailyNotesConfigPath = ".obsidian/daily-notes.json"

// ResolveDailyNotePath returns vault-relative path for a daily note.
func ResolveDailyNotePath(cfg DailyNoteConfig, date time.Time) string {
	format := cfg.Format
	if format == "" {
		format = "YYYY-MM-DD" // Obsidian default
	}
	filename := formatMomentDate(format, date) + ".md"
	if cfg.Folder == "" {
		return filename
	}
	return strings.TrimSuffix(cfg.Folder, "/") + "/" + filename
}

func formatMomentDate(format string, t time.Time) string {
	tokens := []struct{ moment, goFmt string }{
		{"YYYY", "2006"},
		{"MMMM", "January"},
		{"DDDD", "002"},
		{"dddd", "Monday"},
		{"MMM", "Jan"},
		{"ddd", "Mon"},
		{"YY", "06"},
		{"MM", "01"},
		{"DD", "02"},
		{"HH", "15"},
		{"hh", "03"},
		{"mm", "04"},
		{"ss", "05"},
		{"M", "1"},
		{"D", "2"},
		{"A", "PM"},
		{"a", "pm"},
	}
	var result strings.Builder
	for i := 0; i < len(format); {
		// Escape sequences: [literal text]
		if format[i] == '[' {
			// find matching bracket
			end := strings.Index(format[i+1:], "]")
			if end != -1 {
				result.WriteString(format[i+1 : i+1+end])
				i += end + 2 // skip both brackets
				continue
			}
		}
		matched := false
		for _, tok := range tokens {
			if strings.HasPrefix(format[i:], tok.moment) {
				result.WriteString(t.Format(tok.goFmt))
				i += len(tok.moment)
				matched = true
				break
			}
		}
		if !matched {
			result.WriteByte(format[i])
			i++
		}
	}
	return result.String()
}
