package kafka

import (
	"context"

	client "github.com/segmentio/kafka-go"
)

// Config structure
type Config struct {
	UseTLS  bool `mapstructure:"use_tls"`
	Brokers []string
}

// Producer inferface
type Producer interface {
	Start(ctx context.Context) error
	WriteMessages(context.Context, ...client.Message) error
	Close() error
}

// Consumer interface
type Consumer interface {
	Start(ctx context.Context) error
	SetOffset(offset int64) error
	GetOffset() int64
	GetMessageChan() <-chan client.Message
	CommitMessages(context.Context, ...client.Message)
	Close() error
}
