package rest

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
	
	// Bright colors
	ColorBrightRed    = "\033[91m"
	ColorBrightGreen  = "\033[92m"
	ColorBrightYellow = "\033[93m"
	ColorBrightBlue   = "\033[94m"
	ColorBrightPurple = "\033[95m"
	ColorBrightCyan   = "\033[96m"
	
	// Bold
	ColorBold = "\033[1m"
)

// ColorfulHandler is a custom slog handler that provides colorized console output
type ColorfulHandler struct {
	opts   slog.HandlerOptions
	writer io.Writer
	attrs  []slog.Attr
	groups []string
}

// NewColorfulHandler creates a new colorful console handler
func NewColorfulHandler(w io.Writer, opts *slog.HandlerOptions) *ColorfulHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}
	return &ColorfulHandler{
		opts:   *opts,
		writer: w,
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *ColorfulHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle handles the Record
func (h *ColorfulHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format time
	timeStr := r.Time.Format("15:04:05.000")
	
	// Get level color and text
	levelColor, levelText := h.getLevelColorAndText(r.Level)
	
	// Build the log line
	var buf strings.Builder
	
	// Time (gray)
	buf.WriteString(ColorGray)
	buf.WriteString(timeStr)
	buf.WriteString(ColorReset)
	buf.WriteString(" ")
	
	// Level (colored)
	buf.WriteString(levelColor)
	buf.WriteString(ColorBold)
	buf.WriteString(levelText)
	buf.WriteString(ColorReset)
	buf.WriteString(" ")
	
	// Message (white/colored based on level)
	msgColor := ColorWhite
	if r.Level >= slog.LevelError {
		msgColor = ColorBrightRed
	} else if r.Level >= slog.LevelWarn {
		msgColor = ColorBrightYellow
	} else if strings.Contains(r.Message, "HTTP Request") {
		msgColor = ColorBrightCyan
	}
	
	buf.WriteString(msgColor)
	buf.WriteString(r.Message)
	buf.WriteString(ColorReset)
	
	// Add attributes
	if r.NumAttrs() > 0 {
		buf.WriteString(" ")
		h.appendAttrs(&buf, r)
	}
	
	// Add source location if enabled
	if h.opts.AddSource {
		if fs := runtime.CallersFrames([]uintptr{r.PC}); fs != nil {
			f, _ := fs.Next()
			buf.WriteString(ColorGray)
			buf.WriteString(fmt.Sprintf(" (%s:%d)", f.File, f.Line))
			buf.WriteString(ColorReset)
		}
	}
	
	buf.WriteString("\n")
	
	_, err := h.writer.Write([]byte(buf.String()))
	return err
}

// WithAttrs returns a new handler with the given attributes
func (h *ColorfulHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColorfulHandler{
		opts:   h.opts,
		writer: h.writer,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

// WithGroup returns a new handler with the given group
func (h *ColorfulHandler) WithGroup(name string) slog.Handler {
	return &ColorfulHandler{
		opts:   h.opts,
		writer: h.writer,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

// getLevelColorAndText returns the color and text for a given level
func (h *ColorfulHandler) getLevelColorAndText(level slog.Level) (string, string) {
	switch {
	case level >= slog.LevelError:
		return ColorBrightRed, "ERROR"
	case level >= slog.LevelWarn:
		return ColorBrightYellow, "WARN "
	case level >= slog.LevelInfo:
		return ColorBrightGreen, "INFO "
	default:
		return ColorGray, "DEBUG"
	}
}

// appendAttrs formats and appends attributes to the buffer
func (h *ColorfulHandler) appendAttrs(buf *strings.Builder, r slog.Record) {
	first := true
	
	r.Attrs(func(a slog.Attr) bool {
		if !first {
			buf.WriteString(" ")
		}
		first = false
		
		h.appendAttr(buf, a)
		return true
	})
}

// appendAttr formats and appends a single attribute
func (h *ColorfulHandler) appendAttr(buf *strings.Builder, attr slog.Attr) {
	// Special handling for certain attribute keys
	switch attr.Key {
	case "error":
		buf.WriteString(ColorBrightRed)
		buf.WriteString("error")
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		buf.WriteString(ColorRed)
		buf.WriteString(fmt.Sprintf("%v", attr.Value.Any()))
		buf.WriteString(ColorReset)
	case "status", "status_code":
		buf.WriteString(ColorCyan)
		buf.WriteString(attr.Key)
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		status := fmt.Sprintf("%v", attr.Value.Any())
		statusColor := h.getStatusColor(status)
		buf.WriteString(statusColor)
		buf.WriteString(status)
		buf.WriteString(ColorReset)
	case "method":
		buf.WriteString(ColorBlue)
		buf.WriteString("method")
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		buf.WriteString(ColorBrightBlue)
		buf.WriteString(fmt.Sprintf("%v", attr.Value.Any()))
		buf.WriteString(ColorReset)
	case "path":
		buf.WriteString(ColorPurple)
		buf.WriteString("path")
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		buf.WriteString(ColorBrightPurple)
		buf.WriteString(fmt.Sprintf("%v", attr.Value.Any()))
		buf.WriteString(ColorReset)
	case "latency":
		buf.WriteString(ColorGreen)
		buf.WriteString("latency")
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		latency := fmt.Sprintf("%v", attr.Value.Any())
		latencyColor := h.getLatencyColor(latency)
		buf.WriteString(latencyColor)
		buf.WriteString(latency)
		buf.WriteString(ColorReset)
	case "project_id":
		buf.WriteString(ColorYellow)
		buf.WriteString("project_id")
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		buf.WriteString(ColorBrightYellow)
		buf.WriteString(fmt.Sprintf("%v", attr.Value.Any()))
		buf.WriteString(ColorReset)
	default:
		// Default attribute formatting
		buf.WriteString(ColorCyan)
		buf.WriteString(attr.Key)
		buf.WriteString(ColorReset)
		buf.WriteString("=")
		
		// Value formatting based on type
		switch v := attr.Value.Any().(type) {
		case string:
			buf.WriteString(ColorWhite)
			buf.WriteString(fmt.Sprintf("%q", v))
			buf.WriteString(ColorReset)
		case int, int32, int64, float32, float64:
			buf.WriteString(ColorBrightCyan)
			buf.WriteString(fmt.Sprintf("%v", v))
			buf.WriteString(ColorReset)
		case bool:
			color := ColorGreen
			if !v {
				color = ColorRed
			}
			buf.WriteString(color)
			buf.WriteString(fmt.Sprintf("%v", v))
			buf.WriteString(ColorReset)
		case time.Duration:
			buf.WriteString(ColorBrightGreen)
			buf.WriteString(v.String())
			buf.WriteString(ColorReset)
		default:
			buf.WriteString(ColorWhite)
			buf.WriteString(fmt.Sprintf("%v", v))
			buf.WriteString(ColorReset)
		}
	}
}

// getStatusColor returns appropriate color for HTTP status codes
func (h *ColorfulHandler) getStatusColor(status string) string {
	switch {
	case strings.HasPrefix(status, "2"):
		return ColorBrightGreen
	case strings.HasPrefix(status, "3"):
		return ColorBrightBlue
	case strings.HasPrefix(status, "4"):
		return ColorBrightYellow
	case strings.HasPrefix(status, "5"):
		return ColorBrightRed
	default:
		return ColorWhite
	}
}

// getLatencyColor returns appropriate color for latency values
func (h *ColorfulHandler) getLatencyColor(latency string) string {
	if strings.Contains(latency, "ms") {
		if strings.Contains(latency, "µs") || (strings.Contains(latency, "ms") && !strings.Contains(latency, "00ms")) {
			return ColorBrightGreen // Fast < 100ms
		}
		return ColorYellow // Medium
	}
	if strings.Contains(latency, "s") {
		return ColorRed // Slow > 1s
	}
	return ColorGreen // Very fast (µs)
}