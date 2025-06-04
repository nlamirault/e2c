// SPDX-FileCopyrightText: Copyright (C) Nicolas Lamirault <nicolas.lamirault@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

// Level represents the log level
type Level string

const (
	// DebugLevel logs debug messages
	DebugLevel Level = "debug"
	// InfoLevel logs informational messages
	InfoLevel Level = "info"
	// WarnLevel logs warning messages
	WarnLevel Level = "warn"
	// ErrorLevel logs error messages
	ErrorLevel Level = "error"
)

// Format represents the log output format
type Format string

const (
	// TextFormat outputs logs in human-readable text format
	TextFormat Format = "text"
	// JSONFormat outputs logs in JSON format
	JSONFormat Format = "json"
)

// String returns the string representation of the format
func (f Format) String() string {
	return string(f)
}

// Valid checks if a format is valid
func (f Format) Valid() bool {
	return f == TextFormat || f == JSONFormat
}

// Config holds the logger configuration
type Config struct {
	// Level is the logging level (debug, info, warn, error)
	Level Level
	// Format is the output format (text, json)
	Format Format
	// Output is the destination for logs (defaults to stdout)
	Output *os.File
	// AddSource adds the source file and line number to log messages
	AddSource bool
}

// NewConfig creates a default logger configuration
func NewConfig() *Config {
	return &Config{
		Level:     InfoLevel,
		Format:    TextFormat,
		Output:    os.Stdout,
		AddSource: true,
	}
}

// New creates a new logger with the given configuration
func New(cfg *Config) *slog.Logger {
	// If no config is provided, use default
	if cfg == nil {
		cfg = NewConfig()
	}

	// Convert string level to slog.Level
	var level slog.Level
	switch cfg.Level {
	case DebugLevel:
		level = slog.LevelDebug
	case InfoLevel:
		level = slog.LevelInfo
	case WarnLevel:
		level = slog.LevelWarn
	case ErrorLevel:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
		fmt.Fprintf(os.Stderr, "Unknown log level %q, defaulting to info\n", cfg.Level)
	}

	// Set up handler based on format
	var handler slog.Handler
	if cfg.Format == JSONFormat {
		handler = slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.AddSource,
		})
	} else {
		handler = tint.NewHandler(cfg.Output, &tint.Options{
			Level:      level,
			AddSource:  cfg.AddSource,
			TimeFormat: time.RFC3339,
		})
	}

	// Create and return logger
	logger := slog.New(handler)
	return logger
}

// ParseLevel converts a string level to a Level
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		fmt.Fprintf(os.Stderr, "Unknown log level %q, defaulting to info\n", level)
		return InfoLevel
	}
}

// ParseFormat converts a string format to a Format
func ParseFormat(format string) Format {
	switch format {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		fmt.Fprintf(os.Stderr, "Unknown log format %q, defaulting to text\n", format)
		return TextFormat
	}
}

// SetAsDefault sets the given logger as the default slog logger
func SetAsDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}
