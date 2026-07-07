package vault

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type DailyNoteConfig struct {
	Folder   string `json:"folder"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

const dailyNotesConfigPath = ".obsidian/daily-notes.json"

// ReadDailyNoteConfig reads the daily notes plugin config.
// Returns zero config (not error) if plugin not configured.
func (v *Vault) ReadDailyNoteConfig() (DailyNoteConfig, error) {
	data, err := v.ReadFile(context.Background(), dailyNotesConfigPath)
	if errors.Is(err, errNotFound) {
		return DailyNoteConfig{}, nil
	}
	if err != nil {
		return DailyNoteConfig{}, err
	}
	var cfg DailyNoteConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DailyNoteConfig{}, err
	}
	return cfg, nil
}

// ResolveDailyNotePath returns vault-relative path for a daily note.
func (v *Vault) ResolveDailyNotePath(cfg DailyNoteConfig, date time.Time) string {
	format := cfg.Format
	if format == "" {
		format = "YYY-MM-DD" // Obsidian default
	}
	filename := formatMomentDate(format, date) + ".md"
	if cfg.Folder == "" {
		return filename
	}
	return strings.TrimSuffix(cfg.Folder, "/") + "/" + filename
}

func formatMomentDate(format string, t time.Time) string {
	replacements := []struct{ moment, go_ string }{
		{"YYYY", "2006"},
		{"YY", "06"},
		{"MMMM", "January"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"M", "1"},
		{"DDDD", "002"},
		{"DD", "02"},
		{"D", "2"},
		{"dddd", "Monday"},
		{"ddd", "Mon"},
		{"HH", "15"},
		{"hh", "03"},
		{"mm", "04"},
		{"ss", "05"},
		{"A", "PM"},
		{"a", "pm"},
	}
	goFmt := format
	for _, r := range replacements {
		goFmt = strings.ReplaceAll(goFmt, r.moment, r.go_)
	}
	return t.Format(goFmt)
}
