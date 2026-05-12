package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/febriW/order-processing/common/models"
)

type OrderOutboxRepository interface {
	ClaimPendingOutboxEvents(limit, maxAttempts int, owner string, claimTTL time.Duration) ([]OrderOutboxEvent, error)
	MarkOutboxEventPublished(id string) error
	MarkOutboxEventRetry(id, reason string, nextAttemptAt time.Time, dead bool) error
}

type OutboxRelay struct {
	repo        OrderOutboxRepository
	publisher   OrderEventPublisher
	interval    time.Duration
	batchSize   int
	maxAttempts int
	baseBackoff time.Duration
	maxBackoff  time.Duration
	claimTTL    time.Duration
	owner       string
}

func NewOutboxRelay(repo OrderOutboxRepository, publisher OrderEventPublisher, interval time.Duration, batchSize int) *OutboxRelay {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 20
	}
	return &OutboxRelay{
		repo:        repo,
		publisher:   publisher,
		interval:    interval,
		batchSize:   batchSize,
		maxAttempts: 8,
		baseBackoff: 2 * time.Second,
		maxBackoff:  2 * time.Minute,
		claimTTL:    30 * time.Second,
		owner:       envOrDefault("HOSTNAME", "order-service"),
	}
}

func (r *OutboxRelay) Start() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for range ticker.C {
		r.flushOnce()
	}
}

func (r *OutboxRelay) flushOnce() {
	events, err := r.repo.ClaimPendingOutboxEvents(r.batchSize, r.maxAttempts, r.owner, r.claimTTL)
	if err != nil {
		logJSON("error", "outbox_claim_failed", map[string]any{
			"error": err.Error(),
			"owner": r.owner,
		})
		return
	}
	if len(events) > 0 {
		logJSON("info", "outbox_claimed_batch", map[string]any{
			"owner": r.owner,
			"count": len(events),
		})
	}

	for _, event := range events {
		if event.EventType != orderCreatedKey {
			r.scheduleRetry(event, "unsupported event type")
			continue
		}

		var order models.Order
		if err := json.Unmarshal(event.Payload, &order); err != nil {
			r.scheduleRetry(event, "invalid payload")
			continue
		}

		if err := r.publisher.PublishOrderCreated(order, EventPublishMeta{
			EventID:       event.ID,
			TraceID:       event.TraceID,
			CorrelationID: event.CorrelationID,
		}); err != nil {
			logJSON("error", "outbox_publish_failed", map[string]any{
				"event_id":       event.ID,
				"event_type":     event.EventType,
				"attempt":        event.Attempts + 1,
				"trace_id":       event.TraceID,
				"correlation_id": event.CorrelationID,
				"error":          err.Error(),
			})
			r.scheduleRetry(event, err.Error())
			continue
		}
		logJSON("info", "outbox_publish_succeeded", map[string]any{
			"event_id":       event.ID,
			"event_type":     event.EventType,
			"attempt":        event.Attempts + 1,
			"trace_id":       event.TraceID,
			"correlation_id": event.CorrelationID,
		})

		if err := r.repo.MarkOutboxEventPublished(event.ID); err != nil {
			logJSON("error", "outbox_mark_published_failed", map[string]any{
				"event_id":       event.ID,
				"trace_id":       event.TraceID,
				"correlation_id": event.CorrelationID,
				"error":          err.Error(),
			})
		}
	}
}

func (r *OutboxRelay) scheduleRetry(event OrderOutboxEvent, reason string) {
	nextAttempt := time.Now().UTC().Add(r.computeBackoff(event.Attempts + 1))
	dead := event.Attempts+1 >= r.maxAttempts
	if err := r.repo.MarkOutboxEventRetry(event.ID, reason, nextAttempt, dead); err != nil {
		logJSON("error", "outbox_mark_retry_failed", map[string]any{
			"event_id":       event.ID,
			"trace_id":       event.TraceID,
			"correlation_id": event.CorrelationID,
			"error":          err.Error(),
		})
		return
	}
	level := "warn"
	eventName := "outbox_retry_scheduled"
	if dead {
		level = "error"
		eventName = "outbox_moved_dead"
	}
	logJSON(level, eventName, map[string]any{
		"event_id":        event.ID,
		"attempt":         event.Attempts + 1,
		"max_attempts":    r.maxAttempts,
		"next_attempt_at": nextAttempt.Format(time.RFC3339Nano),
		"trace_id":        event.TraceID,
		"correlation_id":  event.CorrelationID,
		"reason":          reason,
	})
}

func (r *OutboxRelay) computeBackoff(attempt int) time.Duration {
	backoff := r.baseBackoff
	for i := 1; i < attempt; i++ {
		backoff = backoff * 2
		if backoff >= r.maxBackoff {
			backoff = r.maxBackoff
			break
		}
	}

	jitterRange := int64(backoff / 4)
	if jitterRange <= 0 {
		return backoff
	}
	jitter := time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(jitterRange))
	if os.Getenv("OUTBOX_DISABLE_JITTER") == "1" {
		return backoff
	}
	return backoff + jitter
}
