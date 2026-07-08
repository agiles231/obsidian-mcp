package vault

import (
	"path/filepath"
	"os"
	"testing"
	"time"
)

func TestResolveDailyNotePath(t *testing.T) {
	// We need a vault inst, but ResolveDailyNotePath doesn't use vault fields.
	// Create minimal one.
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	fixed := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		cfg DailyNoteConfig
		want string
	}{
		{
			name: "default format, no folder",
			cfg: DailyNoteConfig{},
			want: "2026-07-06.md",
		},
		{
			name: "default format, with folder",
			cfg: DailyNoteConfig{Folder: "Daily"},
			want: "Daily/2026-07-06.md",
		},
		{
			name: "custom format",
			cfg: DailyNoteConfig{Format: "YYYY/MM/DD", Folder: "Journal"},
			want: "Journal/2026/07/06.md",
		},
		{
			name: "folder with trailing slash",
			cfg: DailyNoteConfig{Folder: "Daily/"},
			want: "Daily/2026-07-06.md",
		},
		{
			name: "nested folder",
			cfg: DailyNoteConfig{Folder: "Notes/Daily"},
			want: "Notes/Daily/2026-07-06.md",
		},
		{
			name: "format with weekday",
			cfg: DailyNoteConfig{Format: "YYYY-MM-DD dddd"},
			want: "2026-07-06 Monday.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.ResolveDailyNotePath(tt.cfg, fixed)
			if got != tt.want {
				t.Errorf("ResolveDailyNotePath() = %q, want %q", got, tt.want)
			}
		})
	}

}

func TestReadDailyNoteConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	cfg, err := v.ReadDailyNoteConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cfg != (DailyNoteConfig{}) {
		t.Errorf("expected zero config, got %+v", cfg)
	}
}

func TestReadDailyNoteConfig_Valid(t *testing.T) {
	tests := []struct{
		name string
		fileContents string
		want DailyNoteConfig
	}{
		{
			"full config",
			`{"folder": "Daily", "format": "YYYY-MM-DD", "template": "Templates/Daily"}`,
			DailyNoteConfig{
				Folder: "Daily",
				Format: "YYYY-MM-DD",
				Template: "Templates/Daily",
			},
		},
		{
			"partial config",
			`{"folder": "Journal"}`,
			DailyNoteConfig{
				Folder: "Journal",
				Format: "",
				Template: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			// Create .obsidian/daily-notes.json
			obsDir := filepath.Join(dir, ".obsidian")
			if err := os.Mkdir(obsDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte(tt.fileContents), 0644); err != nil {
				t.Fatal(err)
			}

			v, err := Open(Config{Name: "test", Root: dir})
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			cfg, err := v.ReadDailyNoteConfig()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if cfg.Folder != tt.want.Folder || cfg.Format != tt.want.Format || cfg.Template !=  tt.want.Template {
				t.Errorf("unexpected config: %+v", cfg)
			}
		})
	}
}

func TestReadDailyNoteConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	// Create .obsidian/daily-notes.json
	obsDir := filepath.Join(dir, ".obsidian")
	if err := os.Mkdir(obsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(obsDir, "daily-notes.json"), []byte("{not valid}"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := Open(Config{Name: "test", Root: dir})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	_, err = v.ReadDailyNoteConfig()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

}

func TestFormatMomentDate(t *testing.T) {
	// Fixed time: Monday, July 6, 2026, 14:35:09
	fixed := time.Date(2026, 7,6,14,35,9,0, time.UTC)

	tests := []struct {
		name string
		format string
		want string
	}{
		// Common Obsidian daily note formats
		{"iso_date", "YYYY-MM-DD", "2026-07-06"},
		{"slashes", "YYYY/MM/DD", "2026/07/06"},
		{"us style", "MM-DD-YYYY", "07-06-2026"},

		// Year tokens
		{"full year", "YYYY", "2026"},
		{"short year", "YY", "26"},

		// Month tokens
		{"month padded", "MM", "07"},
		{"month unpadded", "M", "7"},
		{"month full name", "MMMM", "July"},
		{"month abbrev name", "MMM", "Jul"},

		// Day tokens
		{"day padded", "DD", "06"},
		{"day unpadded", "D", "6"},
		{"day of year", "DDDD", "187"},

		// Weekday tokens
		{"weekday ful", "dddd", "Monday"},
		{"weekday abbrev", "ddd", "Mon"},

		// Time tokens
		{"hour 24h", "HH", "14"},
		{"hour 12h", "hh", "02"},
		{"minutes", "mm", "35"},
		{"seconds", "ss", "09"},
		{"am/pm upper", "A", "PM"},
		{"am/pm lower", "a", "pm"},

		// Combined formats
		{"datetime", "YYYY-MM-DD HH:mm", "2026-07-06 14:35"},
		{"with weekday", "dddd, MMMM D, YYYY", "Monday, July 6, 2026"},
		{"time only", "hh:mm A", "02:35 PM"},

		// Edge: no tokens
		{"literal only", "[notes]", "notes"},
		{"mixed literal", "[daily]-YYYY", "daily-2026"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMomentDate(tt.format, fixed)
			if got != tt.want {
				t.Errorf("formatMomentDate(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}
