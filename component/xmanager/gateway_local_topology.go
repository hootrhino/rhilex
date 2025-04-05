package xmanager

import "encoding/json"

// LocalTopology represents the local topology of devices in the system
type DataPoint struct {
	ID          string         `json:"id"`          // Unique identifier for the data point
	Name        string         `json:"name"`        // Name of the data point: "temperature", "humidity", etc.
	Type        string         `json:"type"`        // Modbus Type: "holding", "input", "coil", "discrete"
	Address     int            `json:"address"`     // Modbus Address: 0x0000, 0x0001, etc.
	Quantity    int            `json:"quantity"`    // Number of registers or coils to read
	Endian      string         `json:"endian"`      // Endianess: ABCD or DCBA (Modbus)
	Description string         `json:"description"` // Description of the data point
	Unit        string         `json:"unit"`        // Unit of measurement: "Celsius", "Fahrenheit", etc.
	Properties  map[string]any `json:"properties"`  // Additional properties for the data point: "min", "max", "scale"
	Values      []any          `json:"values"`      // Values read from the data points
}

// Device represents a device in the local topology
type Device struct {
	ID              string         `json:"id"`
	Type            string         `json:"type"`      // Type of the device: "sensor", "actuator", "gateway", etc.
	Protocol        string         `json:"protocol"`  // Communication protocol: "Modbus", "MQTT", etc.
	IP              string         `json:"ip"`        // IP address of the device
	Port            int            `json:"port"`      // Port number of the device
	SlaverId        int            `json:"slaver_id"` // Slaver ID for Modbus devices
	Name            string         `json:"name"`
	Status          string         `json:"status"`
	Location        string         `json:"location"`
	Model           string         `json:"model"`
	Manufacturer    string         `json:"manufacturer"`
	SerialNumber    string         `json:"serial_number"`
	FirmwareVersion string         `json:"firmware_version"`
	SoftwareVersion string         `json:"software_version"`
	Properties      map[string]any `json:"properties"`
	DataPoints      []DataPoint    `json:"data_points"`
	LastSeen        string         `json:"last_seen"`
	LastUpdated     string         `json:"last_updated"`
}

// LocalTopology represents the local topology of devices in the system
type LocalTopology struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
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
func (lt *LocalTopology) RemoveDevice(deviceID string) {
	for i, device := range lt.Devices {
		if device.ID == deviceID {
			lt.Devices = append(lt.Devices[:i], lt.Devices[i+1:]...)
			break
		}
	}
}

// GetDevice retrieves a device from the local topology by ID
func (lt *LocalTopology) GetDevice(deviceID string) *Device {
	for _, device := range lt.Devices {
		if device.ID == deviceID {
			return &device
		}
	}
	return nil
}

// UpdateDevice updates the properties of a device in the local topology
func (lt *LocalTopology) UpdateDevice(deviceID string, updatedDevice Device) {
	for i, device := range lt.Devices {
		if device.ID == deviceID {
			lt.Devices[i] = updatedDevice
			break
		}
	}
}

// GetAllDevices retrieves all devices in the local topology
func (lt *LocalTopology) GetAllDevices() []Device {
	return lt.Devices
}
