package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	traceIDContextKey       contextKey = "trace_id"
	correlationIDContextKey contextKey = "correlation_id"
	requestIDContextKey     contextKey = "request_id"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func requestContextMiddleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := headerOrNew(r, "X-Trace-Id")
			correlationID := headerOrNew(r, "X-Correlation-Id")
			requestID := headerOrNew(r, "X-Request-Id")

			ctx := context.WithValue(r.Context(), traceIDContextKey, traceID)
			ctx = context.WithValue(ctx, correlationIDContextKey, correlationID)
			ctx = context.WithValue(ctx, requestIDContextKey, requestID)

			w.Header().Set("X-Trace-Id", traceID)
			w.Header().Set("X-Correlation-Id", correlationID)
			w.Header().Set("X-Request-Id", requestID)

			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))
			logJSON("info", "http_request_completed", map[string]any{
				"service":        service,
				"method":         r.Method,
				"path":           r.URL.Path,
				"status":         rec.status,
				"duration_ms":    time.Since(start).Milliseconds(),
				"trace_id":       traceID,
				"correlation_id": correlationID,
				"request_id":     requestID,
			})
		})
	}
}

func TraceIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(traceIDContextKey).(string)
	if value == "" {
		return uuid.NewString()
	}
	return value
}

func CorrelationIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDContextKey).(string)
	if value == "" {
		return uuid.NewString()
	}
	return value
}

func headerOrNew(r *http.Request, name string) string {
	if value := r.Header.Get(name); value != "" {
		return value
	}
	return uuid.NewString()
}

func logJSON(level, event string, fields map[string]any) {
	record := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level,
		"event":     event,
	}
	for k, v := range fields {
		record[k] = v
	}
	payload, err := json.Marshal(record)
	if err != nil {
		log.Printf("level=error event=logger_marshal_failed err=%v", err)
		return
	}
	log.Print(string(payload))
}
