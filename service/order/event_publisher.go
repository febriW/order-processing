package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/febriW/order-processing/common/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	orderEventsExchange = "order.events"
	orderCreatedKey     = "order.created"
)

type OrderEventPublisher interface {
	PublishOrderCreated(order models.Order, meta EventPublishMeta) error
}

type RabbitPublisher struct {
	conn *amqp.Connection
}

type EventPublishMeta struct {
	EventID       string
	TraceID       string
	CorrelationID string
}

func NewRabbitPublisher(url string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	return &RabbitPublisher{conn: conn}, nil
}

func (p *RabbitPublisher) PublishOrderCreated(order models.Order, meta EventPublishMeta) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(orderEventsExchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return ch.PublishWithContext(ctx, orderEventsExchange, orderCreatedKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		Timestamp:     time.Now().UTC(),
		MessageId:     meta.EventID,
		CorrelationId: meta.CorrelationID,
		Type:          orderCreatedKey,
		Headers: amqp.Table{
			"trace_id":       meta.TraceID,
			"correlation_id": meta.CorrelationID,
			"event_id":       meta.EventID,
		},
		Body: payload,
	})
}

type noopOrderEventPublisher struct{}

func (noopOrderEventPublisher) PublishOrderCreated(order models.Order, meta EventPublishMeta) error {
	return nil
}
