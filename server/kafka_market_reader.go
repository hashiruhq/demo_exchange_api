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
			log.Info().Str("market", id).Int64("last_offset", lastOffset).Msg("Stopping market processor")
			log.Warn().Str("market", id).Str("termination", "shutdown").Msg("Exit market processor")
			return
		case msg, more := <-msgChan:
			if !more {
				log.Info().Str("market", id).Int64("last_offset", lastOffset).Msg("Stopping market processor")
				log.Warn().Str("market", id).Str("termination", "chan_close").Msg("Exit market processor")
				return
			}

			event := data.Event{}
			event.FromBinary(msg.Value)
			lastOffset = msg.Offset

			switch event.Type {
			case data.EventType_NewTrade:
				{
					// output trade info
					trade := event.GetTrade()
					mta.price.SetUint64(trade.Price).SetScale(market.QuotePrecision)
					mta.volume.SetUint64(trade.Amount).SetScale(market.MarketPrecision)
					mta.quoteVolume.Mul(mta.price, mta.volume).Quantize(market.QuotePrecision)

					price, _ := mta.price.Float64()
					volume, _ := mta.volume.Float64()
					quoteVolume, _ := mta.quoteVolume.Float64()

					log.Info().
						Str("market", market.ID).
						Uint64("seq_id", event.SeqID).
						Float64("price_float", price).
						Float64("volume_float", volume).
						Float64("quoteVolume_float", quoteVolume).
						Str("side", trade.TakerSide.String()).
						Uint64("ask_id", trade.AskID).
						Uint64("ask_owner", trade.AskOwnerID).
						Uint64("bid_id", trade.BidID).
						Uint64("bid_owner", trade.BidOwnerID).
						Msg("New trade")
				}
			case data.EventType_OrderStatusChange:
				{
					order := event.GetOrderStatus()
					mta.price.SetUint64(order.Price).SetScale(market.QuotePrecision)
					price, _ := mta.price.Float64()
					mta.volume.SetUint64(order.Amount).SetScale(market.MarketPrecision)
					amount, _ := mta.volume.Float64()

					fundsPrec := market.QuotePrecision
					if order.Side == data.MarketSide_Sell {
						fundsPrec = market.MarketPrecision
					}
					mta.quoteVolume.SetUint64(order.Funds).SetScale(fundsPrec)
					funds, _ := mta.quoteVolume.Float64()

					mta.volume.SetUint64(order.FilledAmount).SetScale(market.MarketPrecision)
					filledAmount, _ := mta.volume.Float64()

					mta.quoteVolume.SetUint64(order.UsedFunds).SetScale(market.QuotePrecision)
					used, _ := mta.quoteVolume.Float64()

					log.Info().
						Str("market", event.Market).
						Str("order_type", order.Type.String()).
						Uint64("order_id", order.ID).
						Uint64("owner_id", order.OwnerID).
						Str("side", order.Side.String()).
						Str("status", order.Status.String()).
						Float64("amount", amount).
						Float64("price", price).
						Float64("funds", funds).
						Float64("filled_amount", filledAmount).
						Float64("used_funds", used).
						Uint64("seq_id", event.SeqID).
						Msg("New order status")
				}
			case data.EventType_OrderActivated:
				{
					order := event.GetOrderActivation()
					mta.price.SetUint64(order.Price).SetScale(market.QuotePrecision)
					price, _ := mta.price.Float64()
					mta.volume.SetUint64(order.Amount).SetScale(market.MarketPrecision)
					amount, _ := mta.volume.Float64()

					fundsPrec := market.QuotePrecision
					if order.Side == data.MarketSide_Sell {
						fundsPrec = market.MarketPrecision
					}
					mta.quoteVolume.SetUint64(order.Funds).SetScale(fundsPrec)
					funds, _ := mta.quoteVolume.Float64()

					mta.volume.SetUint64(order.FilledAmount).SetScale(market.MarketPrecision)
					filledAmount, _ := mta.volume.Float64()

					mta.quoteVolume.SetUint64(order.UsedFunds).SetScale(market.QuotePrecision)
					used, _ := mta.quoteVolume.Float64()

					log.Info().
						Str("market", event.Market).
						Str("order_type", order.Type.String()).
						Uint64("order_id", order.ID).
						Uint64("owner_id", order.OwnerID).
						Str("side", order.Side.String()).
						Str("status", order.Status.String()).
						Float64("amount", amount).
						Float64("price", price).
						Float64("funds", funds).
						Float64("filled_amount", filledAmount).
						Float64("used_funds", used).
						Uint64("seq_id", event.SeqID).
						Msg("Stop order activated")
				}
			case data.EventType_Error:
				{
					orderError := event.GetError()
					log.Error().
						Str("market", event.Market).
						Str("error_code", orderError.Code.String()).
						Uint64("order_id", orderError.OrderID).
						Uint64("seq_id", event.SeqID).
						Msg("Stop order activated")
				}
			}

			// @todo output a message for the rest of the data types.
		}
	}
}
