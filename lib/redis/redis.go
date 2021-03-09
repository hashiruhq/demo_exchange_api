package redis

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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	radix "github.com/mediocregopher/radix/v3"
)

// Config godoc
type Config struct {
	Host       string
	Username   string
	Password   string
	Port       int
	SSLEnabled bool   `mapstructure:"ssl_enabled"`
	CACert     string `mapstructure:"cacert"`
	PoolSize   int
}

// Client connection to redis server
type Client struct {
	Config Config
	Pool   *radix.Pool
}

// NewClient creates a new client
func NewClient(config Config) *Client {
	return &Client{Config: config}
}

// Connect to a redis server
func (client *Client) Connect() error {
	size := client.Config.PoolSize
	connFunc := func(network, addr string) (radix.Conn, error) {
		opts := []radix.DialOpt{}
		if client.Config.Password != "" {
			opts = append(opts, radix.DialAuthPass(client.Config.Password))
		}
		if client.Config.SSLEnabled {
			if client.Config.CACert != "" {
				roots := x509.NewCertPool()
				ok := roots.AppendCertsFromPEM([]byte(client.Config.CACert))
				if !ok {
					panic("failed to parse root certificate")
				}
				tlsConfig := &tls.Config{RootCAs: roots}
				opts = append(opts, radix.DialUseTLS(tlsConfig))
			} else {
				tlsConfig := &tls.Config{InsecureSkipVerify: true}
				opts = append(opts, radix.DialUseTLS(tlsConfig))
			}
		}
		return radix.Dial(network, addr, opts...)
	}

	defaultPoolOpts := []radix.PoolOpt{
		radix.PoolConnFunc(connFunc),
		radix.PoolOnEmptyCreateAfter(1 * time.Second),
		radix.PoolRefillInterval(1 * time.Second),
		radix.PoolOnFullBuffer((size/3)+1, 1*time.Second),
		radix.PoolPingInterval(5 * time.Second / time.Duration(size+1)),
		radix.PoolPipelineConcurrency(size),
		// NOTE if 150us is changed the benchmarks need to be updated too
		radix.PoolPipelineWindow(150*time.Microsecond, 0),
	}

	pool, err := radix.NewPool("tcp", fmt.Sprintf("%s:%d", client.Config.Host, client.Config.Port), client.Config.PoolSize, defaultPoolOpts...)
	if err != nil {
		return err
	}
	client.Pool = pool
	return nil
}

// Disconnect -- close the connection to the redis instance
func (client *Client) Disconnect() error {
	return client.Pool.Close()
}

// Exec processes a command on the redis server with the given arguments
func (client *Client) Exec(val interface{}, command, key string, args ...interface{}) error {
	return client.Pool.Do(radix.FlatCmd(val, command, key, args...))
}
