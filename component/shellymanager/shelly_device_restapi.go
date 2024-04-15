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
	"encoding/json"
	"fmt"
)

// 获取 Shelly 设备配置信息
// https://shelly-api-docs.shelly.cloud/gen2/ComponentsAndServices/Sys#syssetconfig-example
// "http://%s/rpc/Shelly.GetDeviceInfo"
func GetShellyDeviceInfo(Ip string) (ShellyDeviceInfo, error) {
	var ShellyDeviceInfo ShellyDeviceInfo
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Shelly.GetDeviceInfo", Ip))
	if err != nil {
		return ShellyDeviceInfo, err
	}
	err = json.Unmarshal(respBody, &ShellyDeviceInfo)
	if err != nil {
		return ShellyDeviceInfo, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return ShellyDeviceInfo, nil
}

// https://shelly-api-docs.shelly.cloud/gen2/ComponentsAndServices/Sys#sysgetstatus-example
// http://%s/rpc/Sys.GetStatus
func GetShellyDeviceStatus(Ip string) (ShellyDeviceStatus, error) {
	var ShellyDeviceStatus ShellyDeviceStatus
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Sys.GetStatus", Ip))
	if err != nil {
		return ShellyDeviceStatus, err
	}
	err = json.Unmarshal(respBody, &ShellyDeviceStatus)
	if err != nil {
		return ShellyDeviceStatus, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return ShellyDeviceStatus, nil
}
