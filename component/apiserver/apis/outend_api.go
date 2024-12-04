package apis

import (
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/component/lostcache"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

func InitOutEndRoute() {
	OutEndApi := server.RouteGroup(server.ContextUrl("/outends"))
	{
		OutEndApi.GET(("/detail"), server.AddRoute(OutEndDetail))
		OutEndApi.GET(("/list"), server.AddRoute(OutEnds))
		OutEndApi.POST(("/create"), server.AddRoute(CreateOutEnd))
		OutEndApi.DELETE(("/del"), server.AddRoute(DeleteOutEnd))
		OutEndApi.PUT(("/update"), server.AddRoute(UpdateOutEnd))
		OutEndApi.PUT("/restart", server.AddRoute(RestartOutEnd))
		OutEndApi.GET("/outendErrMsg", server.AddRoute(GetOutendErrorMsg))
	}

}
func OutEnds(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	if uuid == "" {
		outends := []typex.OutEnd{}
		for _, mOut := range service.AllMOutEnd() {
			outEnd := ruleEngine.GetOutEnd(mOut.UUID)
			if outEnd == nil {
				tOut := typex.OutEnd{}
				tOut.UUID = mOut.UUID
				tOut.Name = mOut.Name
				tOut.Type = typex.TargetType(mOut.Type)
				tOut.Description = mOut.Description
				tOut.Config = mOut.GetConfig()
				tOut.State = typex.SOURCE_STOP
				outends = append(outends, tOut)
			}
			if outEnd != nil {
				outEnd.State = outEnd.Target.Status()
				outends = append(outends, *outEnd)
			}
		}
		c.JSON(common.HTTP_OK, common.OkWithData(outends))
		return
	}
	mOut, err := service.GetMOutEndWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	outEnd := ruleEngine.GetOutEnd(mOut.UUID)
	if outEnd == nil {
		// 如果内存里面没有就给安排一个死设备
		tOut := typex.OutEnd{}
		tOut.UUID = mOut.UUID
		tOut.Name = mOut.Name
		tOut.Type = typex.TargetType(mOut.Type)
		tOut.Description = mOut.Description
		tOut.Config = mOut.GetConfig()
		tOut.State = typex.SOURCE_STOP
		c.JSON(common.HTTP_OK, common.OkWithData(tOut))
		return
	}
	outEnd.State = outEnd.Target.Status()
	c.JSON(common.HTTP_OK, common.OkWithData(outEnd))
}

// Get all outends
func OutEndDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	mOut, err := service.GetMOutEndWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400EmptyObj(err))
		return
	}
	outEnd := ruleEngine.GetOutEnd(mOut.UUID)
	if outEnd == nil {
		// 如果内存里面没有就给安排一个死设备
		tOutEnd := new(typex.OutEnd)
		tOutEnd.UUID = mOut.UUID
		tOutEnd.Name = mOut.Name
		tOutEnd.Type = typex.TargetType(mOut.Type)
		tOutEnd.Description = mOut.Description
		tOutEnd.Config = mOut.GetConfig()
		tOutEnd.State = typex.SOURCE_STOP
		c.JSON(common.HTTP_OK, common.OkWithData(tOutEnd))
		return
	}
	outEnd.State = outEnd.Target.Status()
	c.JSON(common.HTTP_OK, common.OkWithData(outEnd))
}

// Delete outEnd by UUID
func DeleteOutEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	_, err := service.GetMOutEndWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	if err := service.DeleteMOutEnd(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	old := ruleEngine.GetOutEnd(uuid)
	if old != nil {
		if old.Target.Status() == typex.SOURCE_UP {
			old.Target.Details().State = typex.SOURCE_STOP
			old.Target.Stop()
		}
	}
	ruleEngine.RemoveOutEnd(uuid)
	lostcache.DeleteLostDataTable(uuid)
	c.JSON(common.HTTP_OK, common.Ok())
}

// Create or Update OutEnd
func CreateOutEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID        string                 `json:"uuid"` // 如果空串就是新建, 非空就是更新
		Type        string                 `json:"type" binding:"required"`
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}
	form := Form{}

	if err0 := c.ShouldBindJSON(&form); err0 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err0))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	configJson, err1 := json.Marshal(form.Config)
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	if err := ruleEngine.CheckTargetType(typex.TargetType(form.Type)); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	newUUID := utils.OutUuid()
	if err := service.InsertMOutEnd(&model.MOutEnd{
		UUID:        newUUID,
		Type:        form.Type,
		Name:        form.Name,
		Description: form.Description,
		Config:      string(configJson),
	}); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := server.LoadNewestOutEnd(newUUID, ruleEngine); err != nil {
		c.JSON(common.HTTP_OK, common.OkWithMsg(err.Error()))
		return
	}
	lostcache.CreateLostDataTable(newUUID)
	c.JSON(common.HTTP_OK, common.Ok())

}
func RestartOutEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	err := ruleEngine.RestartOutEnd(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

// 更新
func UpdateOutEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID        string                 `json:"uuid"` // 如果空串就是新建, 非空就是更新
		Type        string                 `json:"type" binding:"required"`
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}
	form := Form{}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	configJson, err := json.Marshal(form.Config)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	// 更新的时候从数据库往外面拿
	OutEnd, err := service.GetMOutEndWithUUID(form.UUID)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	if err := service.UpdateMOutEnd(OutEnd.UUID, &model.MOutEnd{
		UUID:        form.UUID,
		Type:        form.Type,
		Name:        form.Name,
		Description: form.Description,
		Config:      string(configJson),
	}); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	if err := server.LoadNewestOutEnd(form.UUID, ruleEngine); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 获取设备挂了的异常信息
* __DefaultRuleEngine：用于RHILEX内部存储一些KV键值对
 */
func GetOutendErrorMsg(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	Slot := intercache.GetSlot("__DefaultRuleEngine")
	if Slot != nil {
		CacheValue, ok := Slot[uuid]
		if ok {
			c.JSON(common.HTTP_OK, common.OkWithData(CacheValue.ErrMsg))
			return
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithData("--"))
}
