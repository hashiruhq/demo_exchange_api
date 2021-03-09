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
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	client "github.com/segmentio/kafka-go"
)

type kafkaConsumer struct {
	brokers  []string
	topic    string
	inputs   chan client.Message
	consumer *client.Reader
	once     sync.Once
}

// NewKafkaConsumer return a new Kafka consumer
func NewKafkaConsumer(brokers []string, useTLS bool, topic string, partition int) Consumer {
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
	consumer := client.NewReader(client.ReaderConfig{
		Dialer:    dialer,
		Brokers:   brokers,
		Topic:     topic,
		Partition: partition,
		MinBytes:  10,               // 10KB
		MaxBytes:  10 * 1024 * 1024, // 10MB
	})

	return &kafkaConsumer{
		brokers:  brokers,
		topic:    topic,
		consumer: consumer,
		inputs:   make(chan client.Message, 1),
	}
}

func (conn *kafkaConsumer) SetOffset(offset int64) error {
	return conn.consumer.SetOffset(offset)
}

func (conn *kafkaConsumer) GetOffset() int64 {
	return conn.consumer.Offset()
}

// Start the consumer
func (conn *kafkaConsumer) Start(ctx context.Context) error {
	go conn.handleMessages(ctx)
	return nil
}

// GetMessageChan returns the message channel
func (conn *kafkaConsumer) GetMessageChan() <-chan client.Message {
	return conn.inputs
}

// CommitMessages for the given messages
func (conn *kafkaConsumer) CommitMessages(ctx context.Context, msgs ...client.Message) {
	conn.consumer.CommitMessages(ctx, msgs...)
}

// Close the consumer connection
func (conn *kafkaConsumer) Close() (err error) {
	conn.once.Do(func() {
		err = conn.consumer.Close()
		close(conn.inputs)
	})
	return
}

/**
 * Start a supervised reader connection that automatically retries in case of an error and exists in context exit
 */
func (conn *kafkaConsumer) superviseReadingMessages(ctx context.Context) {
	for {
		err := conn.handleMessages(ctx)
		if err == io.EOF {
			// context exited or stream ended
			log.Info().
				Err(err).
				Str("section", "kafka").
				Str("topic", conn.topic).
				Msg("Context exited of message stream ended. Closing consumer")
			break
		}
		log.Error().
			Err(err).
			Str("section", "kafka").
			Str("topic", conn.topic).
			Msg("Unable to read message from reader connection. Closing consumer")
		break
	}
	// closing the connection
	conn.Close()
}

func (conn *kafkaConsumer) handleMessages(ctx context.Context) error {
	for {
		msg, err := conn.consumer.ReadMessage(ctx)
		// context exited or stream ended
		if err == io.EOF {
			return err
		}
		// handle any other kind of error... maybe move it in the supervisor
		if err != nil {
			kafkaErr, ok := err.(client.Error)
			if ok && kafkaErr.Temporary() {
				log.Warn().
					Err(kafkaErr).
					Str("section", "kafka").
					Str("topic", conn.topic).
					Bool("temp", true).
					Msg("Unable to read message from reader connection. Retrying in 1 second")
				// wait some time before retrying
				time.Sleep(time.Second)
				continue
			} else {
				log.Error().
					Err(err).
					Str("section", "kafka").
					Str("topic", conn.topic).
					Bool("temp", false).
					Msg("Unable to read message from kafka server. Exiting message reader loop.")
				return err
			}
		}
		// send the message to the channel for processing
		conn.inputs <- msg
	}
}
