package utils

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/hootrhino/rhilex/ossupport"
	"github.com/hootrhino/rhilex/periphery"
)

/*
*
* 磁盘
*
 */
type DiskUsage struct {
	DeviceID  string  `json:"deviceID"`
	FreeSpace float64 `json:"freeSpace"`
	Size      float64 `json:"size"`
}

func (m DiskUsage) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

/*
*
* CPU
*
 */
type CpuUsage struct {
	Name  string `json:"name"`
	Usage uint64 `json:"usage"`
}

func (m CpuUsage) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

/*
*
* 网卡
*
 */
type NetworkInterfaceUsage struct {
	Name                string
	CurrentBandwidth    uint64
	BytesTotalPerSec    uint64
	BytesReceivedPerSec uint64
	BytesSentPerSec     uint64
	PacketsPerSec       uint64
}

func (m NetworkInterfaceUsage) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

/*
*
* 系统设备
*
 */
type SystemDevices struct {
	Uarts  []string `json:"uarts"`
	Videos []string `json:"videos"`
	Audios []string `json:"audios"`
}

func (m SystemDevices) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

/*
*
* 展示硬件信息
*
 */
func ShowGGpuAndCpuInfo() {
	if runtime.GOARCH == "amd64" {
		if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
			gpus, err1 := periphery.GetGpuInfoWithNvidiaSmi()
			if err1 == nil {
				for _, gpu := range gpus {
					fmt.Println("* Current GPU Device:", gpu.Name)
				}
			} else {
				fmt.Println("* GPU Device Not Found")
			}
			if runtime.GOOS == "linux" {
				cpu, err2 := periphery.GetLinuxCPUName()
				if err2 == nil {
					fmt.Println("* Current CPU Device:", cpu)
				} else {
					fmt.Println("* CPU Detail Not Found")
				}
			}
			if runtime.GOOS == "windows" {
				cpu, err2 := periphery.GetWindowsCPUName()
				if err2 == nil {
					fmt.Println("* Current CPU Device:", cpu)
				} else {
					fmt.Println("* CPU Detail Not Found")

				}
			}
			fmt.Println()
		}
	}

}

/**
 * print all ips
 *
 */
func ShowIpAddress() {
	ipList, err := ossupport.GetAllIps()
	if err != nil {
		fmt.Println("x Get Ip Failed")
	} else {
		fmt.Printf("* Quick Access Address:\n")
		for _, ip := range ipList {
			fmt.Printf("-- %s\n", ip)
		}
		println()
	}
}
