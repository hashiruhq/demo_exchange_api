package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"around25.com/exchange/demo_api/conv"
	"around25.com/exchange/demo_api/data"
	"around25.com/exchange/demo_api/model"
	"github.com/ericlagergren/decimal"
	"github.com/gin-gonic/gin"
	kafkaGo "github.com/segmentio/kafka-go"
)

// AddOrderRoutes godoc
func (srv *server) AddOrderRoutes(r *gin.Engine) {
	group := r.Group("/order")
	{
		group.POST("/:market_id", srv.GetActiveMarket("market_id"), srv.OrderCreate)
		group.DELETE("/:market_id", srv.GetActiveMarket("market_id"), srv.OrderCancel)
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
		if market, ok := marketMap[symbol]; ok {
			c.Set("data_market", &market)
			c.Next()
		} else {
			c.AbortWithStatusJSON(404, map[string]string{"error": "Invalid or inactive market"})
		}
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
	orderType := data.OrderType(data.OrderType_value[c.PostForm("type")])
	side := data.MarketSide(data.MarketSide_value[c.PostForm("side")])
	stop := data.StopLoss(data.StopLoss_value[c.PostForm("stop")])
	id := getPostAsInt(c, "id", 0)
	amount := c.PostForm("amount")
	price := c.PostForm("price")
	stopPrice := c.PostForm("stop_price")
	balance := c.PostForm("balance")
	userID := getPostAsInt(c, "user_id", 0)

	// create order in database and then publish it on apache kafka
	order, err := srv.publishOrder(
		context.TODO(),
		uint64(userID),
		market,
		uint64(id),
		side,
		orderType,
		amount,
		price,
		stop,
		stopPrice,
		balance,
	)
	if err != nil {
		_ = c.Error(err)
		abortWithError(c, 400, err.Error())
		return
	}
	c.JSON(201, order)
}

func (srv *server) OrderCancel(c *gin.Context) {
	iMarket, _ := c.Get("data_market")
	market := iMarket.(*model.Market)
	id := getPostAsInt(c, "id", 0)
	orderType := data.OrderType(data.OrderType_value[c.PostForm("type")])
	side := data.MarketSide(data.MarketSide_value[c.PostForm("side")])
	stop := data.StopLoss(data.StopLoss_value[c.PostForm("stop")])
	price := c.PostForm("price")
	stopPrice := c.PostForm("stop_price")
	userID := getPostAsInt(c, "user_id", 0)

	err := srv.publishCancelOrder(market, uint64(id), data.OrderType(orderType), data.MarketSide(side), price, data.StopLoss(stop), stopPrice, uint64(userID))
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
func (srv *server) publishOrder(
	ctx context.Context,
	userID uint64,
	market *model.Market,
	id uint64,
	side data.MarketSide,
	orderType data.OrderType,
	amount,
	price string,
	stop data.StopLoss,
	stopPrice,
	balance string,
) (*data.Order, error) {
	amountInUnits := conv.ToUnits(amount, uint8(market.MarketPrecision))
	priceInUnits := conv.ToUnits(price, uint8(market.QuotePrecision))
	stopPriceInUnits := conv.ToUnits(stopPrice, uint8(market.QuotePrecision))

	amountAsDecimal, ok := new(decimal.Big).SetString(amount)
	if !ok {
		return nil, errors.New("Invalid amount provided")
	}
	priceAsDecimal, ok := new(decimal.Big).SetString(price)
	if !ok {
		return nil, errors.New("Invalid price provided")
	}

	lockedFunds := new(decimal.Big)

	dBalance, ok := new(decimal.Big).SetString(balance)
	if !ok {
		return nil, errors.New("Invalid balance provided")
	}
	funds := new(decimal.Big).Copy(dBalance)
	if side == data.MarketSide_Sell {
		lockedFunds = new(decimal.Big).Copy(amountAsDecimal)
	} else {
		switch orderType {
		case data.OrderType_Limit:
			lockedFunds = new(decimal.Big).Mul(priceAsDecimal, amountAsDecimal)
		case data.OrderType_Market:
			lockedFunds = new(decimal.Big).Copy(funds)
		}
	}

	fundsInUnits := uint64(0)
	if side == data.MarketSide_Buy {
		fundsInUnits = conv.ToUnits(fmt.Sprintf("%f", lockedFunds), uint8(market.QuotePrecision))
	} else {
		fundsInUnits = conv.ToUnits(fmt.Sprintf("%f", lockedFunds), uint8(market.MarketPrecision))
	}

	// publish order on the registry
	orderEvent := data.Order{
		ID:        id,
		EventType: data.CommandType_NewOrder,
		Side:      side,
		Type:      orderType,
		Stop:      stop,
		Market:    market.ID,
		OwnerID:   userID,
		Amount:    amountInUnits,
		Price:     priceInUnits,
		StopPrice: stopPriceInUnits,
		Funds:     fundsInUnits,
	}
	bytes, err := orderEvent.ToBinary()
	if err != nil {
		return nil, err
	}
	err = srv.publishers[market.ID].WriteMessages(context.TODO(), kafkaGo.Message{Value: bytes})
	return &orderEvent, err
}

// Cancel an existing order
func (srv *server) publishCancelOrder(
	market *model.Market,
	id uint64,
	orderType data.OrderType,
	side data.MarketSide,
	price string,
	stop data.StopLoss,
	stopPrice string,
	userID uint64,
) error {
	priceInUnits := conv.ToUnits(price, uint8(market.QuotePrecision))
	stopPriceInUnits := conv.ToUnits(stopPrice, uint8(market.QuotePrecision))
	// publish order on the registry
	orderEvent := data.Order{
		ID:        id,
		EventType: data.CommandType_CancelOrder,
		Side:      side,
		Type:      orderType,
		Stop:      stop,
		Market:    market.ID,
		OwnerID:   userID,
		Price:     priceInUnits,
		StopPrice: stopPriceInUnits,
	}
	bytes, err := orderEvent.ToBinary()
	if err != nil {
		return err
	}
	return srv.publishers[market.ID].WriteMessages(context.TODO(), kafkaGo.Message{Value: bytes})
}
