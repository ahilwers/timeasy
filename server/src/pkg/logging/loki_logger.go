package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
	"timeasy-server/pkg/configuration"
)

// LokiLogger wraps slog with direct Loki HTTP export capabilities
type LokiLogger struct {
	slogHandler slog.Handler
	lokiHandler *LokiHandler
}

// LokiHandler implements slog.Handler and sends logs directly to Loki
type LokiHandler struct {
	endpoint    string
	bearerToken string
	client      *http.Client
	buffer      chan LokiLogEntry
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	level       slog.Level
}

type LokiLogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Attrs     map[string]interface{}
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type LokiRequest struct {
	Streams []LokiStream `json:"streams"`
}

// NewLokiLogger creates a new logger that sends logs to Loki
// while preserving local console output through slog
func NewLokiLogger(config configuration.Configuration, fallbackHandler slog.Handler) (*LokiLogger, error) {
	var combinedHandler slog.Handler

	// Only setup Loki export if endpoint is configured
	if config.LokiEndpoint != "" && config.LokiBearerToken != "" {
		slog.Info("Initializing Loki logging", "endpoint", config.LokiEndpoint)

		logLevel := config.ParseLogLevel()
		lokiHandler, err := NewLokiHandler(config.LokiEndpoint, config.LokiBearerToken, logLevel)
		if err != nil {
			slog.Error("Failed to create Loki handler", "error", err)
			return nil, err
		}

		// Combine both handlers: Loki direct and local console
		combinedHandler = &MultiHandler{
			handlers: []slog.Handler{lokiHandler, fallbackHandler},
		}

		slog.Info("Loki logging initialized successfully")

		return &LokiLogger{
			slogHandler: combinedHandler,
			lokiHandler: lokiHandler,
		}, nil
	} else {
		slog.Info("Loki logging disabled - missing endpoint or bearer token")
		// Just use the fallback handler
		combinedHandler = fallbackHandler

		return &LokiLogger{
			slogHandler: combinedHandler,
			lokiHandler: nil,
		}, nil
	}
}

func NewLokiHandler(endpoint, bearerToken string, level slog.Level) (*LokiHandler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	handler := &LokiHandler{
		endpoint:    endpoint,
		bearerToken: bearerToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		buffer: make(chan LokiLogEntry, 1000),
		ctx:    ctx,
		cancel: cancel,
		level:  level,
	}

	// Start background worker to send logs
	handler.wg.Add(1)
	go handler.worker()

	return handler, nil
}

// Handler returns the slog handler for use with slog.New()
func (l *LokiLogger) Handler() slog.Handler {
	return l.slogHandler
}

func (l *LokiLogger) Shutdown(ctx context.Context) error {
	if l.lokiHandler != nil {
		return l.lokiHandler.Shutdown(ctx)
	}
	return nil
}

func (h *LokiHandler) Shutdown(ctx context.Context) error {
	h.cancel()
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *LokiHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *LokiHandler) Handle(_ context.Context, record slog.Record) error {
	// Convert slog level to Loki level
	var level string
	switch record.Level {
	case slog.LevelDebug:
		level = "debug"
	case slog.LevelInfo:
		level = "info"
	case slog.LevelWarn:
		level = "warning"
	case slog.LevelError:
		level = "error"
	default:
		level = "info"
	}

	// Extract attributes
	attrs := make(map[string]interface{})
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})

	entry := LokiLogEntry{
		Timestamp: record.Time,
		Level:     level,
		Message:   record.Message,
		Attrs:     attrs,
	}

	// Try to send to buffer, drop if full
	select {
	case h.buffer <- entry:
	default:
		// Buffer full, drop log
	}

	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return h
}

// worker processes log entries and sends them to Loki
func (h *LokiHandler) worker() {
	defer h.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := make([]LokiLogEntry, 0, 100)

	for {
		select {
		case <-h.ctx.Done():
			// Send remaining batch before exiting
			if len(batch) > 0 {
				h.sendBatch(batch)
			}
			return
		case entry := <-h.buffer:
			batch = append(batch, entry)
			if len(batch) >= 100 {
				h.sendBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				h.sendBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// sendBatch sends a batch of log entries to Loki
func (h *LokiHandler) sendBatch(entries []LokiLogEntry) {
	if len(entries) == 0 {
		return
	}

	// Convert entries to Loki format
	values := make([][]string, len(entries))
	for i, entry := range entries {
		// Create log line with message and attributes
		logData := map[string]interface{}{
			"msg":   entry.Message,
			"level": entry.Level,
		}

		// Add attributes
		for k, v := range entry.Attrs {
			logData[k] = v
		}

		logJSON, _ := json.Marshal(logData)

		values[i] = []string{
			strconv.FormatInt(entry.Timestamp.UnixNano(), 10),
			string(logJSON),
		}
	}

	stream := LokiStream{
		Stream: map[string]string{
			"service":     "timeasy",
			"application": "timeasy-server",
			"env":         "production",
		},
		Values: values,
	}

	request := LokiRequest{
		Streams: []LokiStream{stream},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(h.ctx, "POST", h.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.bearerToken)

	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		// Log error but don't fail
		return
	}
}

// MultiHandler implements slog.Handler and forwards to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range m.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				// Log the error but continue with other handlers
				slog.Error("Handler failed", "error", err, "handler", handler)
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, handler := range m.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
