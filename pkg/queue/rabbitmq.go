package queue

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type TransactionMessage struct {
	SenderID   uint    `json:"sender_id"`
	ReceiverID uint    `json:"receiver_id"`
	Amount     float64 `json:"amount"`
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	queues := []string{"transaction_queue", "notification_queue", "audit_queue"}
	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare queue %s: %v", queue, err)
		}
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) PublishTransaction(ctx context.Context, msg TransactionMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",
		"transaction_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (r *RabbitMQ) ConsumeTransactions(handler func(TransactionMessage) error) {
	msgs, err := r.channel.Consume(
		"transaction_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	go func() {
		for d := range msgs {
			var msg TransactionMessage
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				d.Reject(false)
				continue
			}

			err = handler(msg)
			if err != nil {
				log.Printf("Error handling message: %v", err)
				d.Reject(false)
			} else {
				d.Ack(false)
			}
		}
	}()
}

func (r *RabbitMQ) PublishNotification(ctx context.Context, userID uint, message string) error {
	body, err := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"message": message,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",
		"notification_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to publish notification: %v", err)
	}

	return nil
}

func (r *RabbitMQ) PublishAuditLog(ctx context.Context, action string, details interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"action":     action,
		"details":    details,
		"timestamp": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %v", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",
		"audit_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		})

	if err != nil {
		return fmt.Errorf("failed to publish audit log: %v", err)
	}

	return nil
} 