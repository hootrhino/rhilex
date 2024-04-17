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

// The following components are available in Shelly Pro 1:
// System
// WiFi
// Ethernet
// Bluetooth Low Energy
// Cloud
// MQTT
// Outbound Websocket
// 2 instances of Input (input:0, input:1)
// 1 instance of Switch (switch:0)
// Up to 10 instances of Script
import (
	"encoding/json"
	"fmt"
)

func GetPro1DeviceInfo(Ip string) (ShellyDeviceInfo, error) {
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

func GetPro1DeviceStatus(Ip string) (ShellyDeviceStatus, error) {
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

type Pro1InputStatus struct {
	ID     int  `json:"id"`
	Status bool `json:"output"`
}

func GetPro1Input1Status(Ip string) (Pro1InputStatus, error) {
	Pro1InputStatus := Pro1InputStatus{}
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Input.GetStatus?id=0", Ip))
	if err != nil {
		return Pro1InputStatus, err
	}
	err = json.Unmarshal(respBody, &Pro1InputStatus)
	if err != nil {
		return Pro1InputStatus, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return Pro1InputStatus, nil
}
func GetPro1Input2Status(Ip string) (Pro1InputStatus, error) {
	Pro1InputStatus := Pro1InputStatus{}
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Input.GetStatus?id=1", Ip))
	if err != nil {
		return Pro1InputStatus, err
	}
	err = json.Unmarshal(respBody, &Pro1InputStatus)
	if err != nil {
		return Pro1InputStatus, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return Pro1InputStatus, nil
}

type Pro1SwitchStatus struct {
	ID          int    `json:"id"`
	Source      string `json:"source"`
	Output      bool   `json:"output"`
	Temperature struct {
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	} `json:"temperature"`
}

func GetPro1Switch1Status(Ip string) (Pro1SwitchStatus, error) {
	Pro1Switch1Status := Pro1SwitchStatus{}
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Switch.GetStatus?id=0", Ip))
	if err != nil {
		return Pro1Switch1Status, err
	}
	err = json.Unmarshal(respBody, &Pro1Switch1Status)
	if err != nil {
		return Pro1Switch1Status, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return Pro1Switch1Status, nil
}

// http://192.168.1.106/rpc/Switch.Toggle?id=0
type ProToggleSwitch1 struct {
	WasOn bool `json:"was_on"`
}

/*
*
* 翻转开关
*
 */
func Pro1ToggleSwitch1(Ip string) (ProToggleSwitch1, error) {
	ProToggleSwitch1 := ProToggleSwitch1{}
	respBody, err := HttpGet(fmt.Sprintf("http://%s/rpc/Switch.Toggle?id=0", Ip))
	if err != nil {
		return ProToggleSwitch1, err
	}
	err = json.Unmarshal(respBody, &ProToggleSwitch1)
	if err != nil {
		return ProToggleSwitch1, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return ProToggleSwitch1, nil
}
