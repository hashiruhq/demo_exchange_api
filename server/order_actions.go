package server

import (
	"context"
	"fmt"
	"strconv"

	"around25.com/exchange/demo_api/conv"
	"around25.com/exchange/demo_api/data"
	"github.com/ericlagergren/decimal"
	"github.com/gin-gonic/gin"
	kafkaGo "github.com/segmentio/kafka-go"
)

// AddOrderRoutes godoc
func (srv *server) AddOrderRoutes(r *gin.Engine) {
	group := r.Group("/order")
	{
		group.POST("/", srv.GetActiveMarket("market_id"), srv.OrderCreate)
		group.DELETE("/", srv.GetActiveMarket("market_id"), srv.OrderCancel)
	}
}

// GetActiveMarket middleware
// - use this to limit requests to an action based on a given param
func (srv *server) GetActiveMarket(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := c.Param(param)
		// TODO Define a method to get this data from a list of active markets
		// that is kept in memory and updated when something changes, effectively
		// eliminating a call to the database for every call
		marketMap := map[string]model.Market{}
		for _, market := range srv.Config.Markets {
			marketMap[market.ID] = market
		}
		if err != nil {
			c.AbortWithStatusJSON(404, map[string]string{"error": "Invalid or inactive market"})
			return
		}
		c.Set("data_market", &market)
		c.Next()
	}
}

func getQueryAsInt(c *gin.Context, name string, def int) int {
	val := c.Query(name)
	if val == "" {
		return def
	}
	param, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return param
}

func getPostAsInt(c *gin.Context, name string, def int) int {
	val := c.PostForm(name)
	if val == "" {
		return def
	}
	param, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return param
}

func abortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, map[string]interface{}{
		"error": message,
	})
}

func (srv *server) OrderCreate(c *gin.Context) {
	iMarket, _ := c.Get("data_market")
	market := iMarket.(*model.Market)
	orderType := getPostAsInt(c, "type", 0)
	side := getPostAsInt(c, "side", 0)
	amount := c.PostForm("amount")
	price := c.PostForm("price")
	stop := getPostAsInt(c, "stop", 0)
	stopPrice := c.PostForm("stop_price")
	balance := c.PostForm("balance")
	userID := getPostAsInt(c, "user_id", 0)

	// create order in database and then publish it on apache kafka
	order, err := srv.publishOrder(ctx, uint64(userID), market, side, orderType, amount, price, stop, stopPrice, balance)
	if err != nil {
		_ = c.Error(err)
		abortWithError(c, 400, err.Error())
		return
	}
	c.JSON(201, order)
}

func (srv *server) OrderCancel(c *gin.Context) {
	iOrder, _ := c.Get("data_order")
	order := iOrder.(*model.Order)
	iMarket, _ := c.Get("data_market")
	market := iMarket.(*model.Market)
	err := srv.publishCancelOrder(market, order)
	if err != nil {
		_ = c.Error(err)
		abortWithError(c, 500, "Unable to cancel order")
		return
	}
	c.JSON(200, map[string]interface{}{
		"success": true,
		"message": "Order successfully cancelled",
	})
}

// publishOrder and send it to the matching engine based on the given fields
func (srv *server) publishOrder(ctx context.Context, userID uint64, market *model.Market, side int, orderType int, amount, price string, stop int, stopPrice, balance string) (*data.Order, error) {
	amountInUnits := conv.ToUnits(amount, uint8(market.MarketPrecision))
	priceInUnits := conv.ToUnits(price, uint8(market.QuotePrecision))
	stopPriceInUnits := conv.ToUnits(stopPrice, uint8(market.QuotePrecision))
	amountAsDecimal := new(decimal.Big)
	amountAsDecimal.SetString(conv.FromUnits(amountInUnits, uint8(market.MarketPrecision)))
	priceAsDecimal := new(decimal.Big)
	priceAsDecimal.SetString(conv.FromUnits(priceInUnits, uint8(market.QuotePrecision)))
	stopPriceAsDecimal := new(decimal.Big)
	stopPriceAsDecimal.SetString(conv.FromUnits(stopPriceInUnits, uint8(market.QuotePrecision)))

	lockedFunds := new(decimal.Big)
	usedFunds := new(decimal.Big)
	filledAmount := new(decimal.Big)
	feeAmount := new(decimal.Big)

	funds = new(decimal.Big).Copy(balance)
	if order.Side == data.MarketSide_Sell {
		lockedFunds = new(decimal.Big).Copy(amountAsDecimal)
		return
	}

	switch order.Type {
	case data.OrderType_Limit:
		lockedFunds = new(decimal.Big).Mul(priceAsDecimal, amountAsDecimal)
	case data.OrderType_Market:
		lockedFunds = new(decimal.Big).Copy(funds)
	}

	fundsInUnits := uint64(0)
	if order.Side == data.MarketSide_Buy {
		fundsInUnits = conv.ToUnits(fmt.Sprintf("%f", lockedFunds), uint8(market.QuotePrecision))
	} else {
		fundsInUnits = conv.ToUnits(fmt.Sprintf("%f", lockedFunds), uint8(market.MarketPrecision))
	}

	// publish order on the registry
	orderEvent := data.Order{
		ID:        order.ID,
		EventType: data.CommandType_NewOrder,
		Side:      order.Side,
		Type:      order.Type,
		Stop:      order.Stop,
		Market:    order.MarketID,
		OwnerID:   order.OwnerID,
		Amount:    amountInUnits,
		Price:     priceInUnits,
		StopPrice: stopPriceInUnits,
		Funds:     fundsInUnits,
	}
	bytes, err := orderEvent.ToBinary()
	if err != nil {
		return nil, err
	}
	err = server.publishers[market.ID].WriteMessages(ctx, kafkaGo.Message{Value: bytes})
	return order, err
}

// Cancel an existing order
func (srv *server) publishCancelOrder(market *model.Market, order *data.Order) error {
	price := fmt.Sprintf("%f", order.Price.V)
	priceInUnits := conv.ToUnits(price, uint8(market.QuotePrecision))
	stopPrice := fmt.Sprintf("%f", order.StopPrice.V)
	stopPriceInUnits := conv.ToUnits(stopPrice, uint8(market.QuotePrecision))
	// publish order on the registry
	orderEvent := data.Order{
		ID:        order.ID,
		EventType: data.CommandType_CancelOrder,
		Side:      order.Side,
		Type:      order.Type,
		Stop:      order.Stop,
		Market:    order.MarketID,
		OwnerID:   order.OwnerID,
		Price:     priceInUnits,
		StopPrice: stopPriceInUnits,
	}
	bytes, err := orderEvent.ToBinary()
	if err != nil {
		return err
	}
	return server.publishers[market.ID].WriteMessages(ctx, kafkaGo.Message{Value: bytes})
}
