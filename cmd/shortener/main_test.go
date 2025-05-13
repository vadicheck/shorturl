package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
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

	if !strings.Contains(output, "Build version: 1.0.0") {
		t.Errorf("expected 'Build version: 1.0.0', got: %s", output)
	}
	if !strings.Contains(output, "Build date: "+time.Now().Format("2006-01-02")) {
		t.Errorf("expected 'Build date: %s', got: %s", time.Now().Format("2006-01-02"), output)
	}
	if !strings.Contains(output, "Build commit: Short URL YP") {
		t.Errorf("expected 'Build commit: Short URL YP', got: %s", output)
	}
}

func TestPrintBuildInfo_WithValues(t *testing.T) {
	oldLogger := slog.Default()
	defer slog.SetDefault(oldLogger)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	buildVersion = "1.2.4"
	buildDate = "2025-04-19"
	buildCommit = "abc123"

	printBuildInfo()

	output := buf.String()

	if !strings.Contains(output, "Build version: 1.2.4") {
		t.Errorf("expected build version in output, got: %s", output)
	}
	if !strings.Contains(output, "Build date: 2025-04-19") {
		t.Errorf("expected build date in output, got: %s", output)
	}
	if !strings.Contains(output, "Build commit: abc123") {
		t.Errorf("expected build commit in output, got: %s", output)
	}
}
