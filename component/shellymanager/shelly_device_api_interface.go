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

// System
// WiFi
// Ethernet
// Bluetooth Low Energy
// Cloud
// MQTT
// Outbound Websocket
// Input (input:0, input:1)
// Switch (switch:0)
// Up to 10 instances of Script
type ShellyDeviceApi interface {
	GetSystem()    // http://IP/rpc/Sys.GetConfig
	SetSystem()    // http://IP/rpc/Sys.SetConfig
	GetWifi()      // http://IP/rpc/WiFi.SetConfig
	SetWifi()      // http://IP/rpc/WiFi.SetConfig
	GetEthernet()  // http://IP/rpc/Eth.SetConfig
	SetEthernet()  // http://IP/rpc/Eth.SetConfig
	GetBluetooth() // http://IP/rpc/BLE.SetConfig
	SetBluetooth() // http://IP/rpc/BLE.SetConfig
	GetCloud()     // http://IP/rpc/Cloud.SetConfig
	SetCloud()     // http://IP/rpc/Cloud.SetConfig
	GetMqtt()      // http://IP/rpc/MQTT.SetConfig
	SetMqtt()      // http://IP/rpc/MQTT.SetConfig
	GetInput()     // http://IP/rpc/Input.SetConfig
	SetInput()     // http://IP/rpc/Input.SetConfig
	GetSwitch()    // http://IP/rpc/Switch.SetConfig
	SetSwitch()    // http://IP/rpc/Switch.SetConfig
}
