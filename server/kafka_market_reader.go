package server

import (
	"context"
	"time"

	"around25.com/exchange/demo_api/data"
	"around25.com/exchange/demo_api/lib/kafka"
	"around25.com/exchange/demo_api/model"
	"github.com/ericlagergren/decimal"
	"github.com/rs/zerolog/log"
)

const maxReaderBufferSize = 500

type ctxReader string
type tradeAmounts struct {
	price       *decimal.Big
	volume      *decimal.Big
	quoteVolume *decimal.Big
}

func (srv *server) StartMarketProcessor(ctx context.Context, market *model.Market) {
	loopCtx := context.WithValue(ctx, ctxReader("market"), market.ID)
	go srv.loopReadMarketEvents(loopCtx, market)
}

func (srv *server) loopReadMarketEvents(ctx context.Context, market *model.Market) {
	id := market.ID
	mta := &tradeAmounts{
		price:       decimal.New(0, 0),
		volume:      decimal.New(0, 0),
		quoteVolume: decimal.New(0, 0),
	}
	mta.price.Context.RoundingMode = decimal.ToZero
	mta.volume.Context.RoundingMode = decimal.ToZero
	mta.quoteVolume.Context.RoundingMode = decimal.ToZero

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	consumer := kafka.NewKafkaConsumer(srv.Config.Kafka.Brokers, srv.Config.Kafka.UseTLS, "engine.events."+id, 0)
	defer consumer.Close()

	var offset int64 = -2

	//@todo load last processed offset

	log.Info().Str("market", id).Int64("offset", offset).Msg("Start market processor")

	if offset < 0 {
		consumer.SetOffset(offset) // start from the meta offset received
	} else {
		consumer.SetOffset(offset + 1) // start from the next unread offset
	}

	consumer.Start(context.Background())
	msgChan := consumer.GetMessageChan()

	lastOffset := offset

	for {
		select {
		case <-ctx.Done():
			log.Warn().Str("market", id).Str("termination", "shutdown").Msg("Exit market processor")
			return
		case msg, more := <-msgChan:
			if !more {
				log.Warn().Str("market", id).Str("termination", "chan_close").Msg("Exit market processor")
				return
			}

			event := data.Event{}
			event.FromBinary(msg.Value)

			// output trade info
			trade := event.GetTrade()
			if trade != nil {
				mta.price.SetUint64(trade.Price).SetScale(market.QuotePrecision)
				mta.volume.SetUint64(trade.Amount).SetScale(market.MarketPrecision)
				mta.quoteVolume.Mul(mta.price, mta.volume).Quantize(market.QuotePrecision)

				price, _ := mta.price.Float64()
				volume, _ := mta.volume.Float64()
				quoteVolume, _ := mta.quoteVolume.Float64()

				log.Info().
					Str("market", market.ID).
					Int64("seq_id", int64(trade.GetSeqID())).
					Float64("price_float", price).
					Float64("volume_float", volume).
					Float64("quoteVolume_float", quoteVolume).
					Str("side", trade.TakerSide.String()).
					Uint64("ask_id", trade.AskID).
					Uint64("ask_owner", trade.AskOwnerID).
					Uint64("bid_id", trade.BidID).
					Uint64("bid_owner", trade.BidOwnerID).
					Msg("New trade")
				lastOffset = msg.Offset
				continue
			}
			// @todo output a message for the rest of the data types.
		}
	}
	log.Info().Str("market", id).Int64("last_offset", lastOffset).Msg("Stopping market processor")
}
