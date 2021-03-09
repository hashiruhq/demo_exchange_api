package config

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
	"around25.com/exchange/demo_api/lib/kafka"
	"around25.com/exchange/demo_api/model"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config structure
type Config struct {
	Server  ServerConfig
	Kafka   kafka.Config
	Markets []model.Market
}

// ServerConfig structure
type ServerConfig struct {
	Monitoring MonitoringConfig
	API        APIConfig `mapstructure:"api"`
	Exchange   string
}

// APIConfig structure
type APIConfig struct {
	Port   int
	Health bool
	Cors   bool
}

// MonitoringConfig structure
type MonitoringConfig struct {
	Enabled bool
	Host    string
	Port    string
}

// LoadConfig Load server configuration from the yaml file
func LoadConfig(viperConf *viper.Viper) Config {
	var config Config

	err := viperConf.Unmarshal(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to decode config into struct")
	}
	return config
}

// OpenConfig open config file
func OpenConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(".config")
	viper.SetEnvPrefix("CFG")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatal().Err(err).Msg("Unable to read configuration file")
	}
}
