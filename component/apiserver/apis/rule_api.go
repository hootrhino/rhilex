package apis

import (
	"fmt"
	"strings"

	"github.com/hootrhino/beautiful-lua-go/parse"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/component/interqueue"
	"github.com/hootrhino/rhilex/component/luaruntime"
	"github.com/hootrhino/rhilex/glogger"
	transceiver "github.com/hootrhino/rhilex/transceiver"

	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"

	"github.com/gin-gonic/gin"
)

func InitRulesRoute() {
	rulesApi := server.RouteGroup(server.ContextUrl("/rules"))
	{
		rulesApi.POST(("/create"), server.AddRoute(CreateRule))
		rulesApi.PUT(("/update"), server.AddRoute(UpdateRule))
		rulesApi.DELETE(("/del"), server.AddRoute(DeleteRule))
		rulesApi.GET(("/list"), server.AddRoute(Rules))
		rulesApi.GET(("/detail"), server.AddRoute(RuleDetail))
		rulesApi.POST(("/test"), server.AddRoute(TestRulesCallback))
		rulesApi.GET(("/byInend"), server.AddRoute(ListByInend))
		rulesApi.GET(("/byDevice"), server.AddRoute(ListByDevice))
		rulesApi.GET(("/getCanUsedResources"), server.AddRoute(GetAllResources))
		rulesApi.POST(("/formatLua"), server.AddRoute(FormatLua))

	}
}

type ruleVo struct {
	UUID        string   `json:"uuid"`
	FromSource  []string `json:"fromSource"`
	FromDevice  []string `json:"fromDevice"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Status      int      `json:"status"`
	Expression  string   `json:"expression"`
	Description string   `json:"description"`
	Actions     string   `json:"actions"`
	Success     string   `json:"success"`
	Failed      string   `json:"failed"`
}

func RuleDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	rule, err := service.GetMRuleWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400EmptyObj(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(ruleVo{
		UUID:        rule.UUID,
		Name:        rule.Name,
		Status:      1,
		Description: rule.Description,
		FromSource:  []string{rule.SourceId},
		FromDevice:  []string{rule.DeviceId},
		Success:     rule.Success,
		Failed:      rule.Failed,
		Actions:     rule.Actions,
	}))
}

// Get all rules
func Rules(c *gin.Context, ruleEngine typex.Rhilex) {
	DataList := []ruleVo{}
	allRules, _ := service.GetAllMRule()
	for _, rule := range allRules {
		DataList = append(DataList, ruleVo{
			UUID:        rule.UUID,
			Name:        rule.Name,
			Status:      1,
			Description: rule.Description,
			FromSource:  []string{rule.SourceId},
			FromDevice:  []string{rule.DeviceId},
			Success:     rule.Success,
			Failed:      rule.Failed,
			Actions:     rule.Actions,
		})
	}
	c.JSON(common.HTTP_OK, common.OkWithData(DataList))
}

var __default_success = `function Success() end`
var __default_failed = `function Failed(error) end`

// Create rule
func CreateRule(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		FromSource  []string `json:"fromSource" binding:"required"`
		FromDevice  []string `json:"fromDevice" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Actions     string   `json:"actions"`
	}
	form := Form{
		Type: "lua",
	}

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	if !utils.SContains([]string{"lua"}, form.Type) {
		c.JSON(common.HTTP_OK, common.Error(`rule type must be 'lua' but now is:`+form.Type))
		return
	}
	for _, id := range form.FromSource {
		in, _ := service.GetMInEndWithUUID(id)
		if in == nil {
			c.JSON(common.HTTP_OK, common.Error(`inend not exists: `+id))
			return
		}
	}

	for _, id := range form.FromDevice {
		in, _ := service.GetMDeviceWithUUID(id)
		if in == nil {
			c.JSON(common.HTTP_OK, common.Error(`device not exists: `+id))
			return
		}
	}

	// tmpRule 是一个一次性的临时rule，用来验证规则，这么做主要是为了防止真实Lua Vm 被污染
	tmpRule := typex.NewRule(nil, "_", "_", "_", "_", "_",
		__default_success, form.Actions, __default_failed)
	if err := luaruntime.VerifyLuaSyntax(tmpRule); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	rule := typex.NewLuaRule(
		ruleEngine,
		utils.RuleUuid(),
		form.Name,
		form.Description,
		func(s []string) string {
			if len(s) > 0 {
				return s[0]
			}
			return ""
		}(form.FromSource),
		func(s []string) string {
			if len(s) > 0 {
				return s[0]
			}
			return ""
		}(form.FromDevice),
		__default_success,
		form.Actions,
		__default_failed)
	ruleEngine.RemoveRule(rule.UUID)
	if err := ruleEngine.LoadRule(rule); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	go func() {
		// 更新FromSource RULE到Device表中
		for _, inId := range form.FromSource {
			InEnd, _ := service.GetMInEndWithUUID(inId)
			if InEnd == nil {
				// c.JSON(common.HTTP_OK, common.Error(`inend not exists: `+inId))
				glogger.GLogger.Error(`inend not exists: ` + inId)
				return
			}
			// SaveDB
			//
			mRule := &model.MRule{
				Name:        form.Name,
				UUID:        rule.UUID,
				Description: form.Description,
				SourceId: func(s []string) string {
					if len(s) > 0 {
						return s[0]
					}
					return ""
				}(form.FromSource),
				DeviceId: func(s []string) string {
					if len(s) > 0 {
						return s[0]
					}
					return ""
				}(form.FromDevice),
				Success: __default_success,
				Failed:  __default_failed,
				Actions: form.Actions,
			}
			// 去重旧的
			ruleMap := map[string]string{}
			for _, rule := range InEnd.BindRules {
				ruleMap[rule] = rule
			}
			// 追加新的ID
			ruleMap[inId] = mRule.UUID
			// 最后ID列表
			BindRules := []string{}
			for _, iid := range ruleMap {
				BindRules = append(BindRules, iid)
			}
			InEnd.BindRules = BindRules
			if err := service.UpdateMInEnd(InEnd.UUID, &model.MInEnd{
				BindRules: BindRules,
			}); err != nil {
				glogger.GLogger.Error(err)
				return
			}

			if err := service.InsertMRule(mRule); err != nil {
				glogger.GLogger.Error(err)
				return
			}
			// LoadNewest!!!
			if err := server.LoadNewestInEnd(inId, ruleEngine); err != nil {
				glogger.GLogger.Error(err)
				return
			}

		}

		// FromDevice
		for _, devId := range form.FromDevice {
			Device, _ := service.GetMDeviceWithUUID(devId)
			if Device == nil {
				glogger.GLogger.Error(`device not exists: ` + devId)
				return
			}
			// 去重旧的
			ruleMap := map[string]string{}
			for _, rule := range Device.BindRules {
				ruleMap[rule] = rule // for ["", "", "" ....]
			}
			// SaveDB
			mRule := &model.MRule{
				Name:        form.Name,
				UUID:        rule.UUID,
				Description: form.Description,
				SourceId: func(s []string) string {
					if len(s) > 0 {
						return s[0]
					}
					return ""
				}(form.FromSource),
				DeviceId: func(s []string) string {
					if len(s) > 0 {
						return s[0]
					}
					return ""
				}(form.FromDevice),
				Success: __default_success,
				Failed:  __default_failed,
				Actions: form.Actions,
			}
			// 追加新的ID
			ruleMap[devId] = mRule.UUID // append New Rule UUID
			// 最后ID列表
			BindRules := []string{}
			for _, iid := range ruleMap {
				BindRules = append(BindRules, iid)
			}
			Device.BindRules = BindRules
			if err := service.UpdateDevice(Device.UUID, &model.MDevice{
				BindRules: BindRules,
			}); err != nil {
				glogger.GLogger.Error((err))
				return
			}

			if err := service.InsertMRule(mRule); err != nil {
				glogger.GLogger.Error((err))
				return
			}
			// LoadNewest!!!
			if err := server.LoadNewestDevice(devId, ruleEngine); err != nil {
				glogger.GLogger.Error((err))
				return
			}

		}
	}()

	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* Update
*
 */
func UpdateRule(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID        string   `json:"uuid" binding:"required"` // 如果空串就是新建，非空就是更新
		FromSource  []string `json:"fromSource" binding:"required"`
		FromDevice  []string `json:"fromDevice" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Actions     string   `json:"actions"`
	}
	form := Form{Type: "lua"}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	// tmpRule 是一个一次性的临时rule，用来验证规则，这么做主要是为了防止真实Lua Vm 被污染
	tmpRule := typex.NewRule(nil, "_", "_", "_", "_", "_",
		__default_success, form.Actions, __default_failed)
	if err := luaruntime.VerifyLuaSyntax(tmpRule); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	go func() {
		for _, id := range form.FromSource {
			in := ruleEngine.GetInEnd(id)
			if in == nil {
				c.JSON(common.HTTP_OK, common.Error(`inend not exists: `+id))
				return
			}
		}
		for _, id := range form.FromDevice {
			in := ruleEngine.GetDevice(id)
			if in == nil {
				c.JSON(common.HTTP_OK, common.Error(`device not exists: `+id))
				return
			}
		}
		OldRule, err := service.GetMRuleWithUUID(form.UUID)
		if err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		rule := typex.NewLuaRule(
			ruleEngine,
			OldRule.UUID,
			OldRule.Name,
			OldRule.Description,
			OldRule.SourceId,
			OldRule.DeviceId,
			OldRule.Success,
			OldRule.Actions,
			OldRule.Failed)
		ruleEngine.RemoveRule(rule.UUID)
		if err := ruleEngine.LoadRule(rule); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		// SaveDB
		//
		if err := service.UpdateMRule(OldRule.UUID, &model.MRule{
			Name:        form.Name,
			Description: form.Description,
			SourceId: func(s []string) string {
				if len(s) > 0 {
					return s[0]
				}
				return ""
			}(form.FromSource),
			DeviceId: func(s []string) string {
				if len(s) > 0 {
					return s[0]
				}
				return ""
			}(form.FromDevice),
			Success: __default_success,
			Failed:  __default_failed,
			Actions: form.Actions,
		}); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}

		// 耗时操作直接后台执行
		if len(form.FromSource) > 0 {
			// 更新FromSource RULE到Device表中
			InEnd, _ := service.GetMInEndWithUUID(form.FromSource[0])
			if InEnd == nil {
				glogger.GLogger.Error((`inend not exists: ` + form.FromSource[0]))
				return
			}
			// 去重旧的
			ruleMap := map[string]string{}
			for _, rule := range InEnd.BindRules {
				ruleMap[rule] = rule
			}
			// 追加新的ID
			ruleMap[OldRule.UUID] = OldRule.UUID // append New Rule UUID
			// 最后ID列表
			BindRules := []string{}
			for _, iid := range ruleMap {
				BindRules = append(BindRules, iid)
			}
			InEnd.BindRules = BindRules
			if err := service.UpdateMInEnd(InEnd.UUID, &model.MInEnd{
				BindRules: BindRules,
			}); err != nil {
				glogger.GLogger.Error((err))
				return
			}
			// LoadNewest!!!
			if err := server.LoadNewestInEnd(InEnd.UUID, ruleEngine); err != nil {
				glogger.GLogger.Error((err))
				return
			}
		}
		// FromDevice
		if len(form.FromDevice) > 0 {
			Device, _ := service.GetMDeviceWithUUID(form.FromDevice[0])
			if Device == nil {
				glogger.GLogger.Error((`device not exists: ` + form.FromDevice[0]))
				return
			}
			// 去重旧的
			ruleMap := map[string]string{}
			for _, rule := range Device.BindRules {
				ruleMap[rule] = rule // for ["", "", "" ....]
			}
			// 追加新的ID
			ruleMap[OldRule.UUID] = OldRule.UUID // append New Rule UUID
			// 最后ID列表
			BindRules := []string{}
			for _, iid := range ruleMap {
				BindRules = append(BindRules, iid)
			}
			Device.BindRules = BindRules
			if err := service.UpdateDevice(Device.UUID, &model.MDevice{
				BindRules: BindRules,
			}); err != nil {
				glogger.GLogger.Error((err))
				return
			}
			// LoadNewest!!!
			if err := server.LoadNewestDevice(Device.UUID, ruleEngine); err != nil {
				glogger.GLogger.Error((err))
				return
			}
		}
	}()
	c.JSON(common.HTTP_OK, common.Ok())
}

// Delete rule by UUID
func DeleteRule(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	mRule, err0 := service.GetMRule(uuid)
	if err0 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err0))
		return
	}
	if len(mRule.SourceId) > 0 {
		InEnd, _ := service.GetMInEndWithUUID(mRule.SourceId)
		if InEnd == nil {
			c.JSON(common.HTTP_OK, common.Error(`inend not exists: `+mRule.SourceId))
			return
		}
		// 去重旧的
		InEndBindRules := []string{}
		for _, iid := range InEnd.BindRules {
			if iid != mRule.UUID {
				InEndBindRules = append(InEndBindRules, iid)
			}
		}
		if err := service.UpdateMInEnd(InEnd.UUID, &model.MInEnd{
			BindRules: InEndBindRules,
		}); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}

		if err := server.LoadNewestInEnd(mRule.SourceId, ruleEngine); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}

	}

	// FromDevice
	if len(mRule.DeviceId) > 0 {
		Device, _ := service.GetMDeviceWithUUID(mRule.DeviceId)
		if Device == nil {
			c.JSON(common.HTTP_OK, common.Error(`device not exists: `+mRule.DeviceId))
			return
		}
		// 去重旧的
		DeviceBindRules := []string{}
		for _, iid := range Device.BindRules {
			if iid != mRule.UUID {
				DeviceBindRules = append(DeviceBindRules, iid)
			}
		}
		if err := service.UpdateDevice(Device.UUID, &model.MDevice{
			BindRules: DeviceBindRules,
		}); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}
		//
		// 内存里的数据刷新完了以后更新数据库，最后重启
		//
		if err := server.LoadNewestDevice(mRule.DeviceId, ruleEngine); err != nil {
			c.JSON(common.HTTP_OK, common.Error400(err))
			return
		}

	}

	if err := service.DeleteMRule(mRule.UUID); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	ruleEngine.RemoveRule(mRule.UUID)
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 验证lua语法
*
 */
func ValidateLuaSyntax(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		FromSource []string `json:"fromSource" binding:"required"`
		FromDevice []string `json:"fromDevice" binding:"required"`
		Actions    string   `json:"actions" binding:"required"`
		Success    string   `json:"success" binding:"required"`
		Failed     string   `json:"failed" binding:"required"`
	}
	form := Form{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	tmpRule := typex.NewRule(
		nil, // 不需要该字段
		"",  // 不需要该字段
		"",  // 不需要该字段
		"",  // 不需要该字段
		func(s []string) string {
			if len(s) > 0 {
				return s[0]
			}
			return ""
		}(form.FromSource),
		func(s []string) string {
			if len(s) > 0 {
				return s[0]
			}
			return ""
		}(form.FromDevice),
		__default_success,
		form.Actions,
		__default_failed)
	if err := luaruntime.VerifyLuaSyntax(tmpRule); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
	} else {
		c.JSON(common.HTTP_OK, common.Ok())
	}

}

/*
*
* 测试规则脚本
*
 */
func TestRulesCallback(c *gin.Context, ruleEngine typex.Rhilex) {
	type Form struct {
		UUID     string `json:"uuid"`
		Type     string `json:"type"`
		TestData string `json:"testData"`
	}
	form := Form{}
	if err := c.BindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if form.Type == "DEVICE" {
		device := ruleEngine.GetDevice(form.UUID)
		if device == nil {
			c.JSON(common.HTTP_OK, common.Error(fmt.Sprintf("'Device' not exists: %v", form.UUID)))
			return
		}
		err1 := interqueue.PushDeviceQueue(device, "::::TEST_RULE::::"+form.TestData)
		if err1 != nil {
			c.JSON(common.HTTP_OK, common.Error400(err1))
			return
		}
		c.JSON(common.HTTP_OK, common.Ok())
		return
	}
	if form.Type == "INEND" {
		inend := ruleEngine.GetInEnd(form.UUID)
		if inend == nil {
			c.JSON(common.HTTP_OK, common.Error(fmt.Sprintf("'InEnd' not exists: %v", form.UUID)))
			return
		}
		err1 := interqueue.PushInQueue(inend, "::::TEST_RULE::::"+form.TestData)
		if err1 != nil {
			c.JSON(common.HTTP_OK, common.Error400(err1))
			return
		}
		c.JSON(common.HTTP_OK, common.Ok())
		return
	}
	c.JSON(common.HTTP_OK, common.Error("Unsupported Test Type:"+form.Type))
}

/*
*
* 根据设备查询其Rules【0.6.4】
*
 */
func ListByDevice(c *gin.Context, ruleEngine typex.Rhilex) {
	deviceId, _ := c.GetQuery("deviceId")
	MDevice, err := service.GetMDeviceWithUUID(deviceId)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	mRules := service.AllMRules() // 这个效率太低了, 后期写个SQL优化一下
	ruleVos := []ruleVo{}
	for _, rule := range mRules {
		if utils.SContains([]string{rule.DeviceId}, MDevice.UUID) {
			ruleVos = append(ruleVos, ruleVo{
				UUID:        rule.UUID,
				FromSource:  []string{rule.SourceId},
				FromDevice:  []string{rule.DeviceId},
				Name:        rule.Name,
				Status:      1,
				Description: rule.Description,
				Actions:     rule.Actions,
				Success:     rule.Success,
				Failed:      rule.Failed,
			})
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithData(ruleVos))

}

/*
*
* 根据输入查询其Rules【0.6.4】
*
 */
func ListByInend(c *gin.Context, ruleEngine typex.Rhilex) {
	inendId, _ := c.GetQuery("inendId")
	MInend, err := service.GetMInEndWithUUID(inendId)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}

	mRules := service.AllMRules() // 这个效率太低了, 后期写个SQL优化一下
	ruleVos := []ruleVo{}
	for _, rule := range mRules {
		if utils.SContains([]string{rule.SourceId}, MInend.UUID) {
			ruleVos = append(ruleVos, ruleVo{
				UUID:        rule.UUID,
				FromSource:  []string{rule.SourceId},
				FromDevice:  []string{rule.DeviceId},
				Name:        rule.Name,
				Status:      1,
				Description: rule.Description,
				Actions:     rule.Actions,
				Success:     rule.Success,
				Failed:      rule.Failed,
			})
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithData(ruleVos))

}

/*
*
* 给前端快速查询当前系统内的资源
*
 */
type RhilexResource struct {
	UUID string `json:"uuid"` // UUID
	Name string `json:"name"` // 名称
}

func GetAllResources(c *gin.Context, ruleEngine typex.Rhilex) {
	MOutEnds := service.AllMOutEnd()
	MDevices := service.AllDevices()
	OutEnds := []RhilexResource{}
	Devices := []RhilexResource{}
	Schemas := []RhilexResource{}
	RfComs := []RhilexResource{}
	for _, v := range MDevices {
		Devices = append(Devices, RhilexResource{
			UUID: v.UUID,
			Name: v.Name,
		})
	}
	for _, v := range MOutEnds {
		OutEnds = append(OutEnds, RhilexResource{
			UUID: v.UUID,
			Name: v.Name,
		})
	}
	MSchemas := service.AllDataSchema()
	for _, v := range MSchemas {
		Schemas = append(Schemas, RhilexResource{
			UUID: v.UUID,
			Name: v.Name,
		})
	}
	for _, v := range transceiver.List() {
		RfComs = append(RfComs, RhilexResource{
			UUID: v.Name,
			Name: v.Name,
		})
	}
	c.JSON(common.HTTP_OK, common.OkWithData(map[string]any{
		"devices": Devices,
		"outends": OutEnds,
		"schemas": Schemas,
		"rfcoms":  RfComs,
	}))
}

/*
*
* 格式化lua
*
 */
type LuaSource struct {
	Source string `json:"source"`
}

func FormatLua(c *gin.Context, ruleEngine typex.Rhilex) {
	form := LuaSource{}
	if err0 := c.ShouldBindJSON(&form); err0 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err0))
		return
	}
	luaChunk, err1 := parse.Parse(strings.NewReader(form.Source), "")
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(LuaSource{
		Source: luaChunk.String(),
	}))
}
