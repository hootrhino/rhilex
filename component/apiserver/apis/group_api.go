package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

func InitGroupRoute() {
	groupApi := server.RouteGroup(server.ContextUrl("/group"))
	{
		groupApi.POST("/create", server.AddRoute(CreateGroup))
		groupApi.PUT("/update", server.AddRoute(UpdateGroup))
		groupApi.GET("/list", server.AddRoute(ListGroup))
		groupApi.GET("/detail", server.AddRoute(GroupDetail))
		groupApi.POST("/bind", server.AddRoute(BindResource))
		groupApi.PUT("/unbind", server.AddRoute(UnBindResource))
		groupApi.DELETE("/del", server.AddRoute(DeleteGroup))
	}
}

type MGenericGroupVo struct {
	UUID   string `json:"uuid" validate:"required"` // 名称
	Name   string `json:"name" validate:"required"` // 名称
	Type   string `json:"type" validate:"required"` // 组的类型, DEVICE: 设备分组
	Parent string `json:"parent"`                   // 上级, 如果是0表示根节点
}
type MGenericGroupRelationVo struct {
	Gid string `json:"gid" validate:"required"` // 分组ID
	Rid string `json:"rid" validate:"required"` // 被绑定方
}

/*
*
* 新建大屏
*
 */
func CreateGroup(c *gin.Context, ruleEngine typex.Rhilex) {
	vvo := MGenericGroupVo{}
	if err := c.ShouldBindJSON(&vvo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if !utils.SContains([]string{"DEVICE", "USER_LUA_TEMPLATE"}, vvo.Type) {
		c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("Invalid group type [%s]", vvo.Type)))
		return
	}
	Model := model.MGenericGroup{
		UUID:   utils.GroupUuid(),
		Name:   vvo.Name,
		Type:   vvo.Type,
		Parent: "0",
	}
	if err := service.InsertGenericGroup(&Model); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(map[string]string{
		"gid": Model.UUID,
	}))

}

/*
*
* 更新大屏
*
 */
func UpdateGroup(c *gin.Context, ruleEngine typex.Rhilex) {
	vvo := MGenericGroupVo{}
	if err := c.ShouldBindJSON(&vvo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Model := model.MGenericGroup{
		UUID:   vvo.UUID,
		Name:   vvo.Name,
		Type:   vvo.Type,
		Parent: "0",
	}
	if err := service.UpdateGenericGroup(&Model); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 删除大屏
*
 */
func DeleteGroup(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	if uuid == "ROOT" {
		c.JSON(common.HTTP_OK, common.Error("Default group can't delete"))
		return
	}
	count, err := service.CheckBindResource(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 删除之前需要判断一下是不是有子元
	if count > 0 {
		msg := fmt.Errorf("group:%s have binding to other resources", uuid)
		c.JSON(common.HTTP_OK, common.Error400(msg))
		return
	}
	if err := service.DeleteGenericGroup(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 大屏列表
*
 */
func ListGroup(c *gin.Context, ruleEngine typex.Rhilex) {
	MGenericGroups := []MGenericGroupVo{}
	for _, vv := range service.AllGenericGroup() {
		MGenericGroups = append(MGenericGroups, MGenericGroupVo{
			UUID:   vv.UUID,
			Name:   vv.Name,
			Type:   vv.Type,
			Parent: vv.Parent,
		})
	}
	c.JSON(common.HTTP_OK, common.OkWithData(MGenericGroups))

}

func ListDeviceGroup(c *gin.Context, ruleEngine typex.Rhilex) {
	MGenericGroups := []MGenericGroupVo{}
	for _, vv := range service.ListByGroupType("DEVICE") {
		MGenericGroups = append(MGenericGroups, MGenericGroupVo{
			UUID:   vv.UUID,
			Name:   vv.Name,
			Type:   vv.Type,
			Parent: vv.Parent,
		})
	}
	c.JSON(common.HTTP_OK, common.OkWithData(MGenericGroups))
}

/*
*
* 大屏详情
*
 */
func GroupDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	mG, err := service.GetGenericGroupWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
	}
	vo := MGenericGroupVo{
		UUID:   mG.UUID,
		Name:   mG.Name,
		Type:   mG.Type,
		Parent: mG.Parent,
	}
	c.JSON(common.HTTP_OK, common.OkWithData(vo))
}

/*
*
* 绑定资源
*
 */
func BindResource(c *gin.Context, ruleEngine typex.Rhilex) {
	type FormDto struct {
		Gid string `json:"gid"`
		Rid string `json:"rid"`
	}
	form := FormDto{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if count, err := service.CheckAlreadyBinding(form.Gid, form.Rid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	} else {
		if count > 0 {
			msg := fmt.Errorf("[%s] already binding to group [%s]", form.Rid, form.Gid)
			c.JSON(common.HTTP_OK, common.Error400(msg))
			return
		}
	}
	if err := service.BindResource(form.Gid, form.Rid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}

/*
*
* 取消绑定
*
 */
func UnBindResource(c *gin.Context, ruleEngine typex.Rhilex) {
	gid, _ := c.GetQuery("gid")
	rid, _ := c.GetQuery("rid")
	if err := service.UnBindResource(gid, rid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())

}
