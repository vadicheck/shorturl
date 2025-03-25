// Package sl provides helper functions for working with errors in slog.
package sl

import "log/slog"

// Err converts an error into a slog.Attr.
//
// This allows for convenient structured logging of errors
// using the standard slog logger.
//
// Example usage:
//
//	logger.Info("An error occurred", sl.Err(err))
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
