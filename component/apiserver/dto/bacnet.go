package dto

var ValidBacnetObjectType = []string{
	"AO",
	"AI",
	"AV",
	"BI",
	"BO",
	"BV",
	"MI",
	"MO",
	"MV",
}

type BacnetDataPointVO struct {
	UUID           string `json:"uuid"`
	DeviceUUID     string `json:"device_uuid"`
	Tag            string `json:"tag"`
	Alias          string `json:"alias"`
	BacnetDeviceId int    `json:"bacnetDeviceId"`
	ObjectType     string `json:"objectType"`
	ObjectId       int    `json:"objectId"`
	ErrMsg         string `json:"errMsg"`        // 运行时数据
	Status         int    `json:"status"`        // 运行时数据
	LastFetchTime  uint64 `json:"lastFetchTime"` // 运行时数据
	Value          string `json:"value"`         // 运行时数据
}

type BacnetDataPointCreateOrUpdate struct {
	UUID           string `json:"uuid"`
	Tag            string `json:"tag"`
	Alias          string `json:"alias"`
	BacnetDeviceId int    `json:"bacnetDeviceId"`
	ObjectType     string `json:"objectType"`
	ObjectId       int    `json:"objectId"`
}
