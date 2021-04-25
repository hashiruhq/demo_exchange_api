package server

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
	"net/http"

	"around25.com/exchange/demo_api/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (srv *server) SetupHTTPServer() {
	r := gin.New()

	// add middlewares
	r.Use(gin.Recovery()) // Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(logger.SetLogger())

	// add cors if enabled
	srv.ApplyCorsRestrictions(r)
	// add health routes if enabled
	srv.AddHealthRoutes(r)

	srv.AddOrderRoutes(r)

	// configure http server
	srv.HTTP = &http.Server{
		Addr:    ":80",
		Handler: r,
	}
}

func (srv *server) StartHTTPServer() {
	// service connections
	if err := srv.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Str("section", "server").Str("action", "init").Msg("Unable to start HTTP server")
	}
}
