package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestPrintBuildInfo_WithDefaults(t *testing.T) {
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	buildVersion = ""
	buildDate = ""
	buildCommit = ""

	printBuildInfo()

	output := buf.String()

	if !strings.Contains(output, "Build version: N/A") {
		t.Errorf("expected 'Build version: N/A', got: %s", output)
	}
	if !strings.Contains(output, "Build date: N/A") {
		t.Errorf("expected 'Build date: N/A', got: %s", output)
	}
	if !strings.Contains(output, "Build commit: N/A") {
		t.Errorf("expected 'Build commit: N/A', got: %s", output)
	}
}

func TestPrintBuildInfo_WithValues(t *testing.T) {
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	buildVersion = "1.2.3"
	buildDate = "2025-04-19"
	buildCommit = "abc123"

	printBuildInfo()

	output := buf.String()

	if !strings.Contains(output, "Build version: 1.2.3") {
		t.Errorf("expected build version in output, got: %s", output)
	}
	if !strings.Contains(output, "Build date: 2025-04-19") {
		t.Errorf("expected build date in output, got: %s", output)
	}
	if !strings.Contains(output, "Build commit: abc123") {
		t.Errorf("expected build commit in output, got: %s", output)
	}
}
