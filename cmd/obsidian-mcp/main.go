package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/agiles231/mcp-stdio-go"
	"github.com/agiles231/obsidian-mcp/internal/tools"
	"github.com/agiles231/obsidian-mcp/internal/vault"
)

func main() {
	var (
		vaultName  = flag.String("vault", "", "logical vault name (required)")
		vaultRoot  = flag.String("root", "", "path to vault root directory (required)")
		readAllow  = flag.String("read-allow", "", "comma-separated read allow globs (empty = all)")
		writeAllow = flag.String("write-allow", "", "comma-separated write allow globs (empty = none)")
		deny       = flag.String("deny", ".obsidian", "comma-separated deny globs")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if *vaultName == "" || *vaultRoot == "" {
		logger.Error("--vault and --root are required")
		os.Exit(1)
	}

	cfg := vault.Config{
		Name:       *vaultName,
		Root:       *vaultRoot,
		ReadAllow:  splitCSV(*readAllow),
		WriteAllow: splitCSV(*writeAllow),
		Deny:       splitCSV(*deny),
		Logger:     logger,
	}

	v, err := vault.Open(cfg)
	if err != nil {
		logger.Error("failed to open vault", "err", err)
		os.Exit(1)
	}

	registry := vault.NewRegistry()
	if err := registry.Register(v, true); err != nil {
		logger.Error("failed to registry vault", "err", err)
	}

	readFile := tools.NewReadFile(registry)
	writeFile := tools.NewWriteFile(registry)
	appendNote := tools.NewAppendNote(registry)
	searchNotes := tools.NewSearchNotes(registry)
	dailyNote := tools.NewDailyNote(registry)
	listObjects := tools.NewListObjects(registry)

	srv := mcp.NewServer("obsidian-mcp", "0.1.0",
		mcp.WithLogger(logger),
	)
	srv.Register(readFile)
	srv.Register(writeFile)
	srv.Register(appendNote)
	srv.Register(searchNotes)
	srv.Register(dailyNote)
	srv.Register(listObjects)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("starting obsidian-mcp", "vault", *vaultName)
	if err := srv.Run(ctx); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}

}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
