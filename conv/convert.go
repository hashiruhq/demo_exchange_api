package conv

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

// ToUnits converts the given price to uint64 units used by the trading engine
func ToUnits(amounts string, precision uint8) uint64 {
	bytes := []byte(amounts)
	size := len(bytes)
	start := false
	pointPos := 0
	var dec uint64
	i := 0
	for i = 0; i < size && (!start || (start && i-pointPos <= int(precision))); i++ {
		if !start && bytes[i] == '.' {
			start = true
			pointPos = i
		} else {
			dec = 10*dec + uint64(bytes[i]-48) // ascii char for 0
		}
	}
	if !start {
		i = 1
	}
	for i-pointPos <= int(precision) {
		dec *= 10
		i++
	}
	return dec
}

// FromUnits converts the given price to uint64 units used by the trading engine
func FromUnits(number uint64, precision uint8) string {
	bytes := []byte{48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}
	i := 0
	for (number != 0 || i < int(precision)) && i <= 28 {
		add := uint8(number % 10)
		number /= 10
		bytes[28-i] = 48 + add
		if i == int(precision)-1 {
			i++
			bytes[28-i] = 46 // . char
		}
		i++
	}
	i--
	if bytes[28-i] == 46 {
		return string(bytes[28-i-1:])
	}

	return string(bytes[28-i:])
}
