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
	"github.com/rs/zerolog/log"

	"around25.com/exchange/demo_api/config"
	"around25.com/exchange/demo_api/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the server`,
	Run: func(cmd *cobra.Command, args []string) {
		// load server configuration from server
		log.Debug().Msg("Loading server configuration")
		if viper.ConfigFileUsed() != "" {
			log.Debug().Str("section", "init").Str("path", viper.ConfigFileUsed()).Msg("Configuration file loaded")
		}
		cfg := config.LoadConfig(viper.GetViper())
		// start a new server
		log.Debug().Str("section", "init").Msg("Starting new server instance")
		srv := server.NewServer(cfg)
		// listen for new messages
		log.Info().Str("section", "init").Msg("Listening for incoming events")
		srv.Start()
	},
}
