// Copyright (C) 2025 wwhai
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

package rhilex

import "encoding/json"

// LocalTopology represents the local topology of devices in the system
type MetricPoint struct {
	UUID         string  `json:"uuid"`         // Unique identifier for the metric point
	Tag          string  `json:"tag"`          // A unique identifier or label for the register
	Alias        string  `json:"alias"`        // A human-readable name or alias for the register
	SlaverId     uint8   `json:"slaverId"`     // ID of the Modbus slave device
	Function     uint8   `json:"function"`     // Modbus function code (e.g., 3 for Read Holding Registers)
	ReadAddress  uint16  `json:"readAddress"`  // Address of the register in the Modbus device
	ReadQuantity uint16  `json:"readQuantity"` // Number of registers to read/write
	DataType     string  `json:"dataType"`     // Data type of the register value (e.g., uint16, int32, float32)
	DataOrder    string  `json:"dataOrder"`    // Byte order for multi-byte values (e.g., ABCD, DCBA)
	BitPosition  uint16  `json:"bitPosition"`  // bit position for bit-level operations (e.g., 0, 1, 2)
	BitMask      uint16  `json:"bitMask"`      // Bitmask for bit-level operations (e.g., 0x01, 0x02)
	Weight       float64 `json:"weight"`       // Scaling factor for the register value
	Frequency    uint64  `json:"frequency"`    // Polling frequency in milliseconds
	Unit         string  `json:"unit"`         // Unit of measurement: "Celsius", "Fahrenheit", etc.
	Status       string  `json:"status"`       // Status of the register: "active", "inactive", etc.
	Values       []any   `json:"values"`       // Values read from the data points
}

// Device represents a device in the local topology
type Device struct {
	UUID          string         `json:"uuid"`           // Unique identifier for the device
	Type          string         `json:"type"`           // Type of the device: "sensor", "actuator", "gateway", etc.
	Name          string         `json:"name"`           // Name of the device
	Protocol      string         `json:"protocol"`       // Communication protocol: "Modbus", "MQTT", etc.
	SlaverAddress string         `json:"slaver_address"` // Slaver address of the device
	Status        string         `json:"status"`         // Status of the device: "online", "offline", etc.
	SerialNumber  string         `json:"serial_number"`  // Serial number of the device
	Properties    map[string]any `json:"properties"`     // Additional properties of the device
	MetricPoints  []MetricPoint  `json:"data_points"`    // Data points for the device
	LastSeen      string         `json:"last_seen"`      // Last seen timestamp of the device
	LastUpdated   string         `json:"last_updated"`   // Last updated timestamp of the device
}

// LocalTopology represents the local topology of devices in the system
type LocalTopology struct {
	ID      string   `json:"id"`      // Unique identifier for the topology
	Name    string   `json:"name"`    // Name of the topology
	Devices []Device `json:"devices"` // List of devices in the topology
}

func (lt LocalTopology) String() string {
	jsonData, _ := json.Marshal(lt)
	return string(jsonData)
}

// Build a new LocalTopology
func NewLocalTopology(id, name string) *LocalTopology {
	return &LocalTopology{
		ID:      id,
		Name:    name,
		Devices: []Device{},
	}
}

// AddDevice adds a new device to the local topology
func (lt *LocalTopology) AddDevice(device Device) {
	lt.Devices = append(lt.Devices, device)
}

// RemoveDevice removes a device from the local topology by ID
func (lt *LocalTopology) RemoveDevice(uuid string) {
	for i, device := range lt.Devices {
		if device.UUID == uuid {
			lt.Devices = append(lt.Devices[:i], lt.Devices[i+1:]...)
			break
		}
	}
}

// GetDevice retrieves a device from the local topology by ID
func (lt *LocalTopology) GetDevice(uuid string) *Device {
	for _, device := range lt.Devices {
		if device.UUID == uuid {
			return &device
		}
	}
	return nil
}

// UpdateDevice updates the properties of a device in the local topology
func (lt *LocalTopology) UpdateDevice(uuid string, updatedDevice Device) {
	for i, device := range lt.Devices {
		if device.UUID == uuid {
			lt.Devices[i] = updatedDevice
			break
		}
	}
}

// GetAllDevices retrieves all devices in the local topology
func (lt *LocalTopology) GetAllDevices() []Device {
	return lt.Devices
}
