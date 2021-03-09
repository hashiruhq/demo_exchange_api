package cmd

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
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LogLevel Flag
var LogLevel = "info"
var LogFormat = "json"
var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "demo_api",
	Short: "Demo API",
	Long: `A very basic exchange API to forward orders to the matching engine. Created by Around25 to support high frequency trading on crypto markets.
	For a complete documentation and available licenses please contact https://around25.com`,
}

func init() {
	// set log level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	initLoggingEnv()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "", "info", "logging level to show (options: debug|info|warn|error|fatal|panic, default: info)")
	rootCmd.PersistentFlags().StringVarP(&LogFormat, "log-format", "", "info", "log format to generate (Options: json|pretty, default: json)")
	viper.SetConfigName(".config")
	viper.AddConfigPath(".")              // First try to load the config from the current directory
	viper.AddConfigPath("$HOME")          // Then try to load it from the HOME directory
	viper.AddConfigPath("/etc/demo_api/") // As a last resort try to load it from the /etc/market-price
}

func initLoggingEnv() {
	// load log level from env by default
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		LogLevel = logLevel
	}
	// load log format from env by default
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat != "" {
		LogFormat = logFormat
	}
}

func initConfig() {
	// Don't forget to read config either from cfgFile, from current directory or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	customizeLogger()
	viper.SetEnvPrefix("CFG")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Can't read configuration file")
	}
}

// Execute the commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}

func customizeLogger() {
	if LogFormat == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	switch LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	gin.SetMode(gin.ReleaseMode)
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
