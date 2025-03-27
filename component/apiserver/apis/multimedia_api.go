// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package apis

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

func InitMultiMediaRoute() {
	Apis := server.RouteGroup(server.ContextUrl("/multimedia"))
	{
		Apis.POST(("/create"), server.AddRoute(CreateMultiMedia))
		Apis.PUT(("/update"), server.AddRoute(UpdateMultiMedia))
		Apis.GET(("/detail"), server.AddRoute(MultiMediaDetail))
		Apis.GET("/list", server.AddRoute(ListMultiMedia))
		Apis.DELETE(("/del"), server.AddRoute(DeleteMultiMedia))
	}
}

type MultimediaConfig struct {
	StreamUrl  string `json:"streamUrl" validate:"required"`
	EnablePush *bool  `json:"enablePush"`
	PushUrl    string `json:"pushUrl"`
	EnableAi   *bool  `json:"enableAi"`
	AiModel    string `json:"aiModel"`
}

func (cfg MultimediaConfig) Validate() error {
	if cfg.StreamUrl == "" {
		return fmt.Errorf("StreamUrl is required")
	}
	if cfg.PushUrl == "" {
		return fmt.Errorf("PushUrl is required")
	}
	return nil
}

// FromString 从字符串解析配置
func (cfg *MultimediaConfig) FromString(s string) {
	json.Unmarshal([]byte(s), cfg)
}

// JsonString 将配置转换为JSON字符串
func (cfg MultimediaConfig) JsonString() string {
	jsonStr, _ := json.Marshal(cfg)
	return string(jsonStr)
}
func (cfg MultimediaConfig) ToMap() map[string]any {
	m := make(map[string]any)
	m["streamUrl"] = cfg.StreamUrl
	m["enablePush"] = cfg.EnablePush
	m["pushUrl"] = cfg.PushUrl
	m["enableAi"] = cfg.EnableAi
	m["aiModel"] = cfg.AiModel
	return m
}

// RTSP推拉流设置参数
type MultiMediaVo struct {
	UUID        string           `json:"uuid"`                      // 如果空串就是新建, 非空就是更新
	Type        string           `json:"type" binding:"required"`   // 类型
	Name        string           `json:"name" binding:"required"`   // 名称
	Status      int              `json:"status"`                    // 状态
	Config      MultimediaConfig `json:"config" binding:"required"` // 配置
	Description string           `json:"description"`
}

// validateMultiMediaVo 验证MultiMediaVo结构体
func validateMultiMediaVo(multiMediaVo MultiMediaVo) error {
	// 验证类型
	if !utils.SContains([]string{"RTSP_STREAM", "RTMP_STREAM"}, multiMediaVo.Type) {
		return fmt.Errorf("Invalid MultiMedia type [%s]", multiMediaVo.Type)
	}
	// 验证StreamUrl
	if _, err := url.ParseRequestURI(multiMediaVo.Config.StreamUrl); err != nil {
		return err
	}
	// 验证PushUrl
	if multiMediaVo.Config.EnablePush != nil && *multiMediaVo.Config.EnablePush {
		if _, err := url.ParseRequestURI(multiMediaVo.Config.PushUrl); err != nil {
			return err
		}
		// 验证AiModel
		if !utils.SContains([]string{"YOLOV8", "FACENET"}, multiMediaVo.Config.AiModel) {
			return fmt.Errorf("Only Support one of YOLOV8 or FACENET")
		}
	}
	return nil
}

// CreateMultiMedia 创建多媒体资源
func CreateMultiMedia(c *gin.Context, ruleEngine typex.Rhilex) {
	var multiMediaVo MultiMediaVo
	if err := c.ShouldBindJSON(&multiMediaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 验证参数
	if err := validateMultiMediaVo(multiMediaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	configJson, err := json.Marshal(multiMediaVo.Config)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Model := model.MMultiMedia{
		UUID:        utils.MultimediaUuid(),
		Name:        multiMediaVo.Name,
		Type:        multiMediaVo.Type,
		Config:      string(configJson),
		Description: multiMediaVo.Description,
	}
	if errSave := service.InsertMultiMedia(&Model); errSave != nil {
		c.JSON(common.HTTP_OK, common.Error400(errSave))
		return
	}

	c.JSON(common.HTTP_OK, common.Ok())
}

// UpdateMultiMedia 更新多媒体资源
func UpdateMultiMedia(c *gin.Context, ruleEngine typex.Rhilex) {
	var multiMediaVo MultiMediaVo
	if err := c.ShouldBindJSON(&multiMediaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 验证参数
	if err := validateMultiMediaVo(multiMediaVo); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	Model := model.MMultiMedia{
		UUID:        multiMediaVo.UUID,
		Name:        multiMediaVo.Name,
		Type:        multiMediaVo.Type,
		Config:      multiMediaVo.Config.JsonString(),
		Description: multiMediaVo.Description,
	}
	if errSave := service.UpdateMultiMedia(&Model); errSave != nil {
		c.JSON(common.HTTP_OK, common.Error400(errSave))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

// MultiMediaDetail 获取多媒体资源详情
func MultiMediaDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, ok := c.GetQuery("uuid")
	if !ok {
		c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("missing uuid")))
		return
	}
	Model, err := service.GetMultiMediaWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 返回结果
	Config := MultimediaConfig{}
	Config.FromString(Model.Config)
	vo := MultiMediaVo{
		UUID:        Model.UUID,
		Name:        Model.Name,
		Type:        Model.Type,
		Config:      Config,
		Description: Model.Description,
	}
	c.JSON(common.HTTP_OK, common.OkWithData(vo))
}

// ListMultiMedia 获取多媒体资源列表
func ListMultiMedia(c *gin.Context, ruleEngine typex.Rhilex) {
	pager, err := service.ReadPageRequest(c)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	count, MMultiMedias, err := service.PageMultiMedia(pager.Current, pager.Size)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	MultiMedias := []MultiMediaVo{}
	for _, Model := range MMultiMedias {
		// 返回结果
		Config := MultimediaConfig{}
		Config.FromString(Model.Config)
		vo := MultiMediaVo{
			UUID:        Model.UUID,
			Name:        Model.Name,
			Type:        Model.Type,
			Config:      Config,
			Description: Model.Description,
		}
		// Worker, _ := multimedia.GetMultimediaResourceDetails(Model.UUID)
		// if Worker == nil {
		// 	vo.Status = int(xmanager.RESOURCE_DOWN)
		// } else {
		// 	Status := Worker.Worker.Status()
		// 	vo.Status = int(Status)
		// }
		MultiMedias = append(MultiMedias, vo)
	}
	Result := service.WrapPageResult(*pager, MultiMedias, count)
	c.JSON(common.HTTP_OK, common.OkWithData(Result))
}

// DeleteMultiMedia 删除多媒体资源
func DeleteMultiMedia(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, ok := c.GetQuery("uuid")
	if !ok {
		c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("missing uuid")))
		return
	}
	if err := service.DeleteMultiMedia(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("delete failed: %v", err)))
		return
	}

	c.JSON(common.HTTP_OK, common.Ok())
}
