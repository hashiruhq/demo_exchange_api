package data

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
	proto "github.com/golang/protobuf/proto"
)

// NewOrder create a new order
func NewOrder(id, price, amount uint64, side MarketSide, category OrderType, eventType CommandType) Order {
	return Order{ID: id, Price: price, Amount: amount, Side: side, Type: category, EventType: eventType}
}

// Valid checks if the order is valid based on the type of the order and the price/amount/funds
func (order *Order) Valid() bool {
	if order.ID == 0 {
		return false
	}
	switch order.EventType {
	case CommandType_NewOrder:
		{
			if order.Stop != StopLoss_None {
				if order.StopPrice == 0 {
					return false
				}
			}
			switch order.Type {
			case OrderType_Limit:
				return order.Price != 0 && order.Amount != 0
			case OrderType_Market:
				return order.Funds != 0 && order.Amount != 0
			}
		}
	case CommandType_CancelOrder:
		{
			if order.Stop != StopLoss_None {
				if order.StopPrice == 0 {
					return false
				}
			}
			if order.Type == OrderType_Limit {
				return order.Price != 0
			}
		}
	}
	return true
}

// Filled checks if the order can be considered filled
func (order *Order) Filled() bool {
	if order.EventType != CommandType_NewOrder {
		return false
	}
	if order.Type == OrderType_Market && (order.Amount == 0 || order.Funds == 0) {
		return true
	}
	return false
}

//***************************
// Interface Implementations
//***************************

// LessThan implementes the skiplist interface
func (order Order) LessThan(other Order) bool {
	return order.Price < other.Price
}

// FromBinary loads an order from a byte array
func (order *Order) FromBinary(msg []byte) error {
	return proto.Unmarshal(msg, order)
}

// ToBinary converts an order to a byte string
func (order *Order) ToBinary() ([]byte, error) {
	return proto.Marshal(order)
}
