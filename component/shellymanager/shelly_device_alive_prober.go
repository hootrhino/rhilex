// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package shellymanager

import (
	"time"

	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils/tinyarp"
)

/*
*
* 如果没有回复就删除
*
 */
func (M *ShellyDeviceRegistry) TestAlive() {
	go func() {
		for {
			select {
			case <-typex.GCTX.Done():
				return
			default:
			}
			for RulexDeviceId, Slot := range M.Slots {
				go func(RulexDeviceId string, Slot map[string]ShellyDevice) {
					for Mac, Device := range Slot {
						if tinyarp.IsValidIP(Device.Ip) {
							_, err := GetShellyDeviceInfo(Device.Ip)
							if err != nil {
								M.DeleteValue(RulexDeviceId, Mac)
								return
							}
						}
					}
				}(RulexDeviceId, Slot)
			}
			time.Sleep(5000 * time.Millisecond)
		}
	}()
}
