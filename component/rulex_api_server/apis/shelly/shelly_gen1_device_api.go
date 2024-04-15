package shelly

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/rulex_api_server/common"
	"github.com/hootrhino/rhilex/component/rulex_api_server/server"
	"github.com/hootrhino/rhilex/component/shellymanager"
	"github.com/hootrhino/rhilex/typex"
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
}

type ShellyDeviceVo struct {
	Ip         string  `json:"ip"` // 扫描出来的IP
	Name       *string `json:"name"`
	ID         string  `json:"id"`
	Mac        string  `json:"mac"`
	Slot       int     `json:"slot"`
	Model      string  `json:"model"`
	Gen        int     `json:"gen"`
	FwID       string  `json:"fw_id"`
	Ver        string  `json:"ver"`
	App        string  `json:"app"`
	AuthEn     bool    `json:"auth_en"`
	AuthDomain *string `json:"auth_domain"`
}

func ScanShellyDevice(c *gin.Context, ruleEngine typex.RuleX) {
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 删除ShellyDevice
*
 */
func DeleteShellyDevice(c *gin.Context, ruleEngine typex.RuleX) {
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* ShellyDevice列表
*
 */
func ListShellyDevice(c *gin.Context, ruleEngine typex.RuleX) {
	ShellyDevices := []ShellyDeviceVo{}
	uuid, _ := c.GetQuery("uuid")
	Slot := shellymanager.GetSlot(uuid)
	for _, ShellyDevice := range Slot {
		ShellyDevices = append(ShellyDevices, ShellyDeviceVo{
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
		})
	}

	c.JSON(common.HTTP_OK, common.OkWithData(ShellyDevices))

}

/*
*
* 获取Shelly设备的当前状态
*
 */
func ShellyDeviceStatus(c *gin.Context, ruleEngine typex.RuleX) {
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
func ShellyDeviceDetail(c *gin.Context, ruleEngine typex.RuleX) {
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
func ScanDevice(c *gin.Context, ruleEngine typex.RuleX) {
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
