package apis

import (
	"fmt"

	"encoding/json"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/intercache"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

func InitInEndRoute() {
	InEndApi := server.RouteGroup(server.ContextUrl("/inends"))
	{
		InEndApi.GET(("/detail"), server.AddRoute(InEndDetail))
		InEndApi.GET(("/list"), server.AddRoute(InEnds))
		InEndApi.POST(("/create"), server.AddRoute(CreateInend))
		InEndApi.DELETE(("/del"), server.AddRoute(DeleteInEnd))
		InEndApi.PUT(("/update"), server.AddRoute(UpdateInend))
		InEndApi.PUT("/restart", server.AddRoute(RestartInEnd))
		InEndApi.GET("/clients", server.AddRoute(GetInEndClients))
		InEndApi.GET("/inendErrMsg", server.AddRoute(GetInendErrorMsg))
	}
}
func InEndDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	Model, err := service.GetMInEndWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400EmptyObj(err))
		return
	}
	inEnd := ruleEngine.GetInEnd(Model.UUID)
	if inEnd == nil {
		tmpInEnd := typex.InEnd{
			UUID:        Model.UUID,
			Type:        typex.InEndType(Model.Type),
			Name:        Model.Name,
			Description: Model.Description,
			BindRules:   map[string]typex.Rule{},
			Config:      Model.GetConfig(),
			State:       typex.SOURCE_STOP,
		}
		c.JSON(common.HTTP_OK, common.OkWithData(tmpInEnd))
		return
	}
	inEnd.State = inEnd.Source.Status()
	c.JSON(common.HTTP_OK, common.OkWithData(inEnd))
}

// Get all inends
func InEnds(c *gin.Context, ruleEngine typex.Rhilex) {

	inEnds := []typex.InEnd{}
	for _, v := range service.AllMInEnd() {
		var inEnd *typex.InEnd
		if inEnd = ruleEngine.GetInEnd(v.UUID); inEnd == nil {
			tmpInEnd := typex.InEnd{
				UUID:        v.UUID,
				Type:        typex.InEndType(v.Type),
				Name:        v.Name,
				Description: v.Description,
				BindRules:   map[string]typex.Rule{},
				Config:      v.GetConfig(),
				State:       typex.SOURCE_STOP,
			}
			inEnds = append(inEnds, tmpInEnd)
		}
		if inEnd != nil {
			inEnd.State = inEnd.Source.Status()
			inEnds = append(inEnds, *inEnd)
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithData(inEnds))
}

// Create or Update InEnd
func CreateInend(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID        string         `json:"uuid"` // 如果空串就是新建, 非空就是更新
		Type        string         `json:"type" binding:"required"`
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Config      map[string]any `json:"config" binding:"required"`
	}
	form := Form{}

	if err0 := c.ShouldBindJSON(&form); err0 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err0))
		return
	}
	configJson, err1 := json.Marshal(form.Config)
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	if err := ruleEngine.CheckSourceType(typex.InEndType(form.Type)); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	isSingle := false
	// 内部消息总线是单例模式
	if form.Type == typex.INTERNAL_EVENT.String() {
		for _, inend := range ruleEngine.AllInEnds() {
			if inend.Type.String() == form.Type {
				isSingle = true
			}
		}
	}
	if isSingle {
		msg := fmt.Errorf("the %s is singleton Source, can not create again", form.Name)
		c.JSON(common.HTTP_OK, common.Error400(msg))
		return
	}
	newUUID := utils.InUuid()

	if err := service.InsertMInEnd(&model.MInEnd{
		UUID:        newUUID,
		Type:        form.Type,
		Name:        form.Name,
		Description: form.Description,
		Config:      string(configJson),
		XDataModels: "[]",
	}); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := server.LoadNewestInEnd(newUUID, ruleEngine); err != nil {
		c.JSON(common.HTTP_OK, common.OkWithMsg(err.Error()))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}
func RestartInEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	err := ruleEngine.RestartInEnd(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 更新输入资源
*
 */
func UpdateInend(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID        string         `json:"uuid"` // 如果空串就是新建, 非空就是更新
		Type        string         `json:"type" binding:"required"`
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Config      map[string]any `json:"config" binding:"required"`
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
	InEnd, err := service.GetMInEndWithUUID(form.UUID)
	if err != nil {
		c.JSON(common.HTTP_OK, err)
		return
	}

	if err := service.UpdateMInEnd(InEnd.UUID, &model.MInEnd{
		UUID:        form.UUID,
		Type:        form.Type,
		Name:        form.Name,
		Description: form.Description,
		Config:      string(configJson),
	}); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := server.LoadNewestInEnd(form.UUID, ruleEngine); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

// Delete inend by UUID
func DeleteInEnd(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	_, err := service.GetMInEndWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if err := service.DeleteMInEnd(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
	} else {
		old := ruleEngine.GetInEnd(uuid)
		if old != nil {
			old.Source.Stop()
			old.Source.Details().State = typex.SOURCE_STOP
		}
		ruleEngine.RemoveInEnd(uuid)
		c.JSON(common.HTTP_OK, common.Ok())
	}
}

type InEndClient struct {
	Ip         string         `json:"ip"`
	Status     bool           `json:"status"`
	Properties map[string]any `json:"properties"`
}

/*
*
* 获取客户端列表[南向资源可能会有一些客户端连接上来]
*
 */
func GetInEndClients(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// deviceUuid, _ := c.GetQuery("device_uuid")
	// clients_registry.get(deviceUuid) -> []InEndClient
	InEndClients := []InEndClient{
		{Ip: "127.0.0.1", Status: true, Properties: map[string]any{
			"observe": true,
			"version": "2.0",
		}},
	}
	Count := int64(0)
	c.JSON(common.HTTP_OK, common.OkWithData(service.WrapPageResult(*pager, InEndClients, Count)))
}

/*
*
* 获取设备挂了的异常信息
* __DefaultRuleEngine：用于RHILEX内部存储一些KV键值对
 */
func GetInendErrorMsg(c *gin.Context, ruleEngine typex.Rhilex) {
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
