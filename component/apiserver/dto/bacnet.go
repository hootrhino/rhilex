package dto

type BacnetDataPointVO struct {
	UUID           string `json:"uuid,omitempty"`
	DeviceUUID     string `json:"device_uuid,omitempty"`
	Tag            string `json:"tag,omitempty"`
	Alias          string `json:"alias,omitempty"`
	BacnetDeviceId *int   `json:"bacnetDeviceId,omitempty"`
	ObjectType     string `json:"objectType,omitempty"`
	ObjectId       *int   `json:"objectId,omitempty"`
	Frequency      int    `json:"frequency,omitempty"`
	ErrMsg         string `json:"errMsg,omitempty"`        // 运行时数据
	Status         int    `json:"status,omitempty"`        // 运行时数据
	LastFetchTime  uint64 `json:"lastFetchTime,omitempty"` // 运行时数据
	Value          string `json:"value,omitempty"`         // 运行时数据
}
