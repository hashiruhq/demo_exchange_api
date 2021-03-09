package model

/*
 * Copyright © 2018-2019 Around25 SRL <office@around25.com>
 *
 * Licensed under the Around25 Wallet License Agreement (the "License");
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
 * @copyright 2018-2019 Around25 SRL <office@around25.com>
 * @license 	EXCHANGE_LICENSE
 */

// Market structure
type Market struct {
	ID               string `mapstructure:"id"`
	MarketPrecision  int    `mapstructure:"market_precision"`
	QuotePrecision   int    `mapstructure:"quote_precision"`
	MarketCoinSymbol string `mapstructure:"market_coin_symbol"`
	QuoteCoinSymbol  string `mapstructure:"quote_coin_symbol"`
}

// GORM Event Handlers
