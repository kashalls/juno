package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// slogFormatter is a middleware.LogFormatter that logs requests through
// slog, matching the rest of the app's structured log output instead of
// chi's default colorized "scheme://host+path" line.
type slogFormatter struct{}

func (slogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &slogLogEntry{
		method:    r.Method,
		path:      r.URL.Path,
		remoteIP:  middleware.GetClientIP(r.Context()),
		requestID: middleware.GetReqID(r.Context()),
	}
}

type slogLogEntry struct {
	method    string
	path      string
	remoteIP  string
	requestID string
}

func (e *slogLogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	slog.Info("request",
		"method", e.method,
		"path", e.path,
		"status", status,
		"bytes", bytes,
		"duration", elapsed,
		"remote_ip", e.remoteIP,
		"request_id", e.requestID,
	)
}

func (e *slogLogEntry) Panic(v interface{}, stack []byte) {
	slog.Error("panic recovered",
		"method", e.method,
		"path", e.path,
		"request_id", e.requestID,
		"err", v,
		"stack", string(stack),
	)
}
