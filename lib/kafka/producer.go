package kafka

/*
 * Copyright © 2006-2019 Around25 SRL <office@around25.com>
 *
 * Licensed under the Around25 Exchange License Agreement (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.around25.com/licenses/EXCHANGE_LICENSE
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Cosmin Harangus <cosmin@around25.com>
 * @copyright 2006-2019 Around25 SRL <office@around25.com>
 * @license 	EXCHANGE_LICENSE
 */

import (
	"context"
	"crypto/tls"
	"time"

	client "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/snappy"
)

// KafkaProducer structure
type kafkaProducer struct {
	producer *client.Writer
	brokers  []string
	ctx      context.Context
}

// NewKafkaProducer returns a new producer
func NewKafkaProducer(brokers []string, useTLS bool, topic string) Producer {
	var dialer *client.Dialer
	if useTLS {
		tlsCfg := &tls.Config{
			//InsecureSkipVerify: true,
		}
		dialer = &client.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS:       tlsCfg,
		}
	}
	producer := client.NewWriter(client.WriterConfig{
		Dialer:           dialer,
		Brokers:          brokers,
		Topic:            topic,
		QueueCapacity:    100,
		BatchSize:        20000,
		BatchTimeout:     time.Duration(100) * time.Millisecond,
		Async:            false,
		CompressionCodec: snappy.NewCompressionCodec(),
	})
	return &kafkaProducer{
		producer: producer,
		brokers:  brokers,
	}
}

// Start the kafka producer
func (conn *kafkaProducer) Start(ctx context.Context) error {
	conn.ctx = ctx
	return nil
}

// Write one or multiple messages to the topic partition
func (conn *kafkaProducer) WriteMessages(ctx context.Context, msgs ...client.Message) error {
	if ctx == nil {
		ctx = conn.ctx
	}
	return conn.producer.WriteMessages(ctx, msgs...)
}

// Get statistics about the producer since the last time it was executed
func (conn *kafkaProducer) Stats() client.WriterStats {
	return conn.producer.Stats()
}

// Close the producer connection
func (conn *kafkaProducer) Close() error {
	return conn.producer.Close()
}
