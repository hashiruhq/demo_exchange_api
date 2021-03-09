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
	// import http profilling when the server profilling configuration is set
	_ "net/http/pprof"

	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"around25.com/exchange/demo_api/config"
	"github.com/rs/zerolog/log"
)

// Server interface
type Server interface {
	Start()
}

type server struct {
	HTTP   *http.Server
	Config config.Config
	ctx    context.Context
	close  context.CancelFunc
}

// NewServer godoc
func NewServer(cfg config.Config) Server {
	ctx, close := context.WithCancel(context.Background())
	return &server{
		Config: cfg,
		ctx:    ctx,
		close:  close,
	}
}

// Start server
func (srv *server) Start() {
	// start the http server
	srv.SetupHTTPServer()
	go srv.StartHTTPServer()

	markets := srv.Config.Markets
	for i := range markets {
		market := markets[i]
		srv.StartMarketProcessor(srv.ctx, &market)
	}

	// stop server of signal
	srv.stopOnSignal(srv.close)
}

func (srv *server) stopOnSignal(close context.CancelFunc) {
	// wait for termination signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGINT)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	sig := <-sigc
	log.Info().Str("section", "server").Str("action", "terminate").Str("signal", sig.String()).Msg("Shutting down services")

	// shutdown http server within 5 seconds
	ctx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()

	if err := srv.HTTP.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Str("section", "server").Str("action", "terminate").Msg("Unable to shutdown HTTP server")
	}
	// close main context
	close()

	// exit program
	os.Exit(0)
}
