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

// FromBinary loads a trade from a byte array
func (trade *Trade) FromBinary(msg []byte) error {
	return proto.Unmarshal(msg, trade)
}

// ToBinary converts a trade to a byte string
func (trade *Trade) ToBinary() ([]byte, error) {
	return proto.Marshal(trade)
}
