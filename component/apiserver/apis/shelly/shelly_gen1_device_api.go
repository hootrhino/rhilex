package shelly

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/shellymanager"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
	"github.com/hootrhino/rhilex/utils/tinyarp"
)

func InitShellyRoute() {
	ShellyApi := server.RouteGroup(server.ContextUrl("/shelly_gen1"))
	{
		ShellyApi.GET("/list", server.AddRoute(ListShellyDevice))
		ShellyApi.GET("/status", server.AddRoute(ShellyDeviceStatus))
		ShellyApi.GET("/detail", server.AddRoute(ShellyDeviceDetail))
		ShellyApi.DELETE("/del", server.AddRoute(DeleteShellyDevice))
		ShellyApi.POST("/scan", server.AddRoute(ScanShellyDevice))
	}
	Pro1Api := server.RouteGroup(server.ContextUrl("/shelly_gen1/pro1"))
	{
		Pro1Api.GET("/switch1/toggle", server.AddRoute(Pro1ToggleSwitch1))
		Pro1Api.POST("/configWebHook", server.AddRoute(Pro1ConfigWebHook))
	}
}

type ShellyDeviceInPortVo struct {
	ID     int  `json:"id"`
	Status bool `json:"status"`
}
type ShellyDeviceOutPortVo struct {
	ID          int    `json:"id"`
	Source      string `json:"source"`
	Output      bool   `json:"output"`
	Temperature struct {
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	} `json:"temperature"`
}
type ShellyDeviceVo struct {
	Ip         string                  `json:"ip"` // 扫描出来的IP
	Name       *string                 `json:"name"`
	ID         string                  `json:"id"`
	Mac        string                  `json:"mac"`
	Slot       int                     `json:"slot"`
	Model      string                  `json:"model"`
	Gen        int                     `json:"gen"`
	FwID       string                  `json:"fw_id"`
	Ver        string                  `json:"ver"`
	App        string                  `json:"app"`
	AuthEn     bool                    `json:"auth_en"`
	AuthDomain *string                 `json:"auth_domain"`
	Input      []ShellyDeviceInPortVo  `json:"input"`
	Switch     []ShellyDeviceOutPortVo `json:"switch"`
}

func ScanShellyDevice(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	tinyarp.AutoRefresh(1000 * time.Second)
	ArpTable := tinyarp.SendArp()
	go func() {
		// 1 将第一次扫出来请求失败的设备拉进黑名单,防止浪费资源
		// 2 已经有在列表里面的就不再扫描
		for Ip, Mac := range ArpTable {
			if shellymanager.Exists(uuid, Ip) {
				continue
			}
			if !tinyarp.IsValidIP(Ip) {
				continue
			}
			DeviceInfo, err := shellymanager.GetShellyDeviceInfo(Ip)
			if err != nil {
				continue
			}
			DeviceInfo.Mac = Mac
			if utils.IsValidMacAddress1(DeviceInfo.Mac) ||
				utils.IsValidMacAddress2(DeviceInfo.Mac) {
				shellymanager.SetValue(uuid, DeviceInfo.Mac, shellymanager.ShellyDevice{
					Ip:         DeviceInfo.Ip,
					Name:       DeviceInfo.Name,
					ID:         DeviceInfo.ID,
					Mac:        DeviceInfo.Mac,
					Slot:       DeviceInfo.Slot,
					Model:      DeviceInfo.Model,
					Gen:        DeviceInfo.Gen,
					FwID:       DeviceInfo.FwID,
					Ver:        DeviceInfo.Ver,
					App:        DeviceInfo.App,
					AuthEn:     DeviceInfo.AuthEn,
					AuthDomain: DeviceInfo.AuthDomain,
				})
			}
		}
	}()
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 删除ShellyDevice
*
 */
func DeleteShellyDevice(c *gin.Context, ruleEngine typex.Rhilex) {
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* ShellyDevice列表
*
 */
func ListShellyDevice(c *gin.Context, ruleEngine typex.Rhilex) {
	ShellyDevices := []ShellyDeviceVo{}
	uuid, _ := c.GetQuery("uuid")
	Slot := shellymanager.GetSlot(uuid)
	for _, ShellyDevice := range Slot {
		NewShellyDevice := ShellyDeviceVo{
			Ip:         ShellyDevice.Ip,
			Name:       ShellyDevice.Name,
			ID:         ShellyDevice.ID,
			Mac:        ShellyDevice.Mac,
			Slot:       ShellyDevice.Slot,
			Model:      ShellyDevice.Model,
			Gen:        ShellyDevice.Gen,
			FwID:       ShellyDevice.FwID,
			Ver:        ShellyDevice.Ver,
			App:        ShellyDevice.App,
			AuthEn:     ShellyDevice.AuthEn,
			AuthDomain: ShellyDevice.AuthDomain,
			Input:      make([]ShellyDeviceInPortVo, 0),
			Switch:     make([]ShellyDeviceOutPortVo, 0),
		}
		if ShellyDevice.App == "Pro1" {
			{
				Pro1InputStatus, err := shellymanager.GetPro1Input1Status(ShellyDevice.Ip)
				if err != nil {
					NewShellyDevice.Input = append(NewShellyDevice.Input, ShellyDeviceInPortVo{
						ID:     Pro1InputStatus.ID,
						Status: Pro1InputStatus.Status,
					})
				} else {
					NewShellyDevice.Input = append(NewShellyDevice.Input, ShellyDeviceInPortVo{
						ID:     0,
						Status: false,
					})
				}
			}
			{
				Pro1InputStatus, err := shellymanager.GetPro1Input2Status(ShellyDevice.Ip)
				if err != nil {
					NewShellyDevice.Input = append(NewShellyDevice.Input, ShellyDeviceInPortVo{
						ID:     Pro1InputStatus.ID,
						Status: Pro1InputStatus.Status,
					})
				} else {
					NewShellyDevice.Input = append(NewShellyDevice.Input, ShellyDeviceInPortVo{
						ID:     1,
						Status: false,
					})
				}
			}
			{
				Pro1Switch1Status, err := shellymanager.GetPro1Switch1Status(ShellyDevice.Ip)
				if err == nil {
					NewShellyDevice.Switch = append(NewShellyDevice.Switch, ShellyDeviceOutPortVo{
						ID:     0,
						Source: Pro1Switch1Status.Source,
						Output: Pro1Switch1Status.Output,
						Temperature: struct {
							TC float64 "json:\"tC\""
							TF float64 "json:\"tF\""
						}{Pro1Switch1Status.Temperature.TC, Pro1Switch1Status.Temperature.TF},
					})
				} else {
					NewShellyDevice.Switch = append(NewShellyDevice.Switch, ShellyDeviceOutPortVo{
						ID:     0,
						Source: "error",
						Output: false,
						Temperature: struct {
							TC float64 "json:\"tC\""
							TF float64 "json:\"tF\""
						}{0, 0},
					})
				}
			}

		}
		ShellyDevices = append(ShellyDevices, NewShellyDevice)
	}
	c.JSON(common.HTTP_OK, common.OkWithData(ShellyDevices))
}

/*
*
* 获取Shelly设备的当前状态
*
 */
func ShellyDeviceStatus(c *gin.Context, ruleEngine typex.Rhilex) {
	ip, _ := c.GetQuery("ip")
	if tinyarp.IsValidIP(ip) {
		ShellyDeviceStatus, err := shellymanager.GetShellyDeviceStatus(ip)
		if err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		c.JSON(common.HTTP_OK, common.OkWithData(ShellyDeviceStatus))
		return
	}
	c.JSON(common.HTTP_OK, common.Error("invalid ip schema:"+ip))
}

/*
*
* ShellyDevice详情
*
 */
func ShellyDeviceDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	ShellyDevice := ShellyDeviceVo{}
	deviceId, _ := c.GetQuery("deviceId")
	mac, _ := c.GetQuery("mac")
	Slot := shellymanager.GetSlot(deviceId)
	if Slot != nil {
		if SDevice, ok := Slot[mac]; ok {
			ShellyDevice.Ip = SDevice.Ip
			ShellyDevice.Name = SDevice.Name
			ShellyDevice.ID = SDevice.ID
			ShellyDevice.Mac = SDevice.Mac
			ShellyDevice.Slot = SDevice.Slot
			ShellyDevice.Model = SDevice.Model
			ShellyDevice.Gen = SDevice.Gen
			ShellyDevice.FwID = SDevice.FwID
			ShellyDevice.Ver = SDevice.Ver
			ShellyDevice.App = SDevice.App
			ShellyDevice.AuthEn = SDevice.AuthEn
			ShellyDevice.AuthDomain = SDevice.AuthDomain
			c.JSON(common.HTTP_OK, common.OkWithData(ShellyDevice))
			return
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithEmpty())
}

/*
* ShellyDeviceRegistry: 绑定到Server的注册中心
* 扫描设备
*
 */
func ScanDevice(c *gin.Context, ruleEngine typex.Rhilex) {
	tinyarp.AutoRefresh(1000 * time.Second)
	ArpTable := tinyarp.SendArp()
	wg := sync.WaitGroup{}
	wg.Add(len(ArpTable))
	for Ip := range ArpTable {
		go func(Ip string) {
			defer wg.Done()
			if tinyarp.IsValidIP(Ip) {
				DeviceInfo, err := shellymanager.GetShellyDeviceInfo(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}
				shellymanager.SetValue(DeviceInfo.Mac, DeviceInfo.Mac, shellymanager.ShellyDevice{
					Ip:         DeviceInfo.Ip,
					Name:       DeviceInfo.Name,
					ID:         DeviceInfo.ID,
					Mac:        DeviceInfo.Mac,
					Slot:       DeviceInfo.Slot,
					Model:      DeviceInfo.Model,
					Gen:        DeviceInfo.Gen,
					FwID:       DeviceInfo.FwID,
					Ver:        DeviceInfo.Ver,
					App:        DeviceInfo.App,
					AuthEn:     DeviceInfo.AuthEn,
					AuthDomain: DeviceInfo.AuthDomain,
				})
			}
		}(Ip)
	}
	wg.Wait()
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 拨动开关
* http://192.168.1.106/rpc/Switch.Toggle?id=0
 */
func Pro1ToggleSwitch1(c *gin.Context, ruleEngine typex.Rhilex) {
	Ip, _ := c.GetQuery("ip")
	ProToggleSwitch1Response, err := shellymanager.Pro1ToggleSwitch1(Ip)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(ProToggleSwitch1Response))
}

/*
*
* 配置WebHook, 自动触发以及一键清空

	{
	    "id": 1,
	    "method": "Webhook.Create",
	    "params": {
	        "name":"PUSH-SWITCH${cid}-EVENT-ON-TO-RHILEX",
	        "cid": "${cid}",
	        "enable": true,
	        "event": "switch.on",
	        "urls": [
	            "http://192.168.1.175:6400?mac=${config.sys.device.mac}&token=shelly&action=switch_on&cid=${cid}"
	        ]
	    }
	}
*/
func Pro1ConfigWebHook(c *gin.Context, ruleEngine typex.Rhilex) {
	opType, _ := c.GetQuery("opType")
	Ip, _ := c.GetQuery("ip")
	if opType == "set_webhook" {
		if err := shellymanager.Pro1CheckWebhook(Ip); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		go func() {
			{
				_, err := shellymanager.Pro1SetSw0OnHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
			{
				_, err := shellymanager.Pro1SetSw0OffHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
			{
				_, err := shellymanager.Pro1SetInput0OnHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
			{
				_, err := shellymanager.Pro1SetInput0OffHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
			{
				_, err := shellymanager.Pro1SetInput1OnHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
			{
				_, err := shellymanager.Pro1SetInput1OffHook(Ip)
				if err != nil {
					c.JSON(common.HTTP_OK, common.Error400(err))
					return
				}

			}
		}()
		c.JSON(common.HTTP_OK, common.Ok())
		return
	}
	if opType == "clear_webhook" {
		if tinyarp.IsValidIP(Ip) {
			err := shellymanager.Pro1ClearWebhook(Ip)
			if err != nil {
				c.JSON(common.HTTP_OK, common.Error400(err))
				return
			}
		} else {
			c.JSON(common.HTTP_OK, common.Error("Invalid Ip:"+Ip))
			return
		}
	}
	c.JSON(common.HTTP_OK, common.Ok())
}
