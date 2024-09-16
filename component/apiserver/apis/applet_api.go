package apis

import (
	"fmt"

	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/component/apiserver/model"
	"github.com/hootrhino/rhilex/component/apiserver/server"
	"github.com/hootrhino/rhilex/component/apiserver/service"

	"github.com/gin-gonic/gin"
	"github.com/hootrhino/rhilex/component/applet"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

func InitAppletRoute() {
	appApi := server.RouteGroup(server.ContextUrl("/app"))
	{
		appApi.GET(("/list"), server.AddRoute(Apps))
		appApi.POST(("/create"), server.AddRoute(CreateApp))
		appApi.PUT(("/update"), server.AddRoute(UpdateApp))
		appApi.DELETE(("/del"), server.AddRoute(RemoveApp))
		appApi.PUT(("/start"), server.AddRoute(StartApp))
		appApi.PUT(("/stop"), server.AddRoute(StopApp))
		appApi.GET(("/detail"), server.AddRoute(AppDetail))
	}
}

/*
*
* 其实这个结构体扮演的角色VO层
*
 */
type AppletDto struct {
	UUID        string `json:"uuid,omitempty"` // 名称
	Name        string `json:"name"`           // 名称
	Version     string `json:"version"`        // 版本号
	AutoStart   *bool  `json:"autoStart"`      // 自动启动
	AppState    int    `json:"appState"`       // 状态: 1 运行中, 0 停止
	Type        string `json:"type"`           // 默认就是lua, 留个扩展以后可能支持别的
	LuaSource   string `json:"luaSource"`      // Lua源码
	Description string `json:"description"`
}

/*
*
* APP 详情
*
 */
func AppDetail(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	// uuid
	appInfo, err1 := service.GetMAppWithUUID(uuid)
	if err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400EmptyObj(err1))
		return
	}
	web_data := AppletDto{
		UUID:      appInfo.UUID,
		Name:      appInfo.Name,
		Version:   appInfo.Version,
		AutoStart: appInfo.AutoStart,
		Type:      "lua",
		AppState: func() int {
			if a := applet.GetApp(appInfo.UUID); a != nil {
				return int(a.AppState)
			}
			return 0
		}(),
		Description: appInfo.Description,
		LuaSource:   appInfo.LuaSource,
	}
	c.JSON(common.HTTP_OK, common.OkWithData(web_data))
}

// 列表
func Apps(c *gin.Context, ruleEngine typex.Rhilex) {
	result := []AppletDto{}
	for _, mApp := range service.AllApp() {
		web_data := AppletDto{
			UUID:      mApp.UUID,
			Name:      mApp.Name,
			Version:   mApp.Version,
			AutoStart: mApp.AutoStart,
			Type:      "lua",
			AppState: func() int {
				if a := applet.GetApp(mApp.UUID); a != nil {
					return int(a.AppState)
				}
				return 0
			}(),
			Description: mApp.Description,
		}
		result = append(result, web_data)
	}
	c.JSON(common.HTTP_OK, common.OkWithData(result))
}

const luaTemplate = `
--
-- Go https://www.hootrhino.com for more tutorials
--
-- APP ID: %s
-- App NAME = "%s"
-- App VERSION = "%s"
-- App DESCRIPTION = "%s"
--
-- Rhilex Main
--
%s
`
const defaultLuaMain = `
function Main(arg)
	Debug("[Hello Rhilex]:" .. time:Time())
	return 0
end
`

func CreateApp(c *gin.Context, ruleEngine typex.Rhilex) {
	form := AppletDto{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	newUUID := utils.AppUuid()
	mAPP := &model.MApplet{
		UUID:    newUUID,
		Name:    form.Name,
		Version: form.Version,
		LuaSource: fmt.Sprintf(luaTemplate,
			newUUID, form.Name, form.Version, form.Description, defaultLuaMain),
		AutoStart:   form.AutoStart,
		Description: form.Description,
	}
	if err := service.InsertApp(mAPP); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 立即加载但是不运行，主要是要加入内存
	newAPP := applet.NewApplication(newUUID, form.Name, form.Version)
	newAPP.AutoStart = *form.AutoStart
	newAPP.Description = form.Description
	if err := applet.LoadApp(newAPP, mAPP.LuaSource); err != nil {
		glogger.GLogger.Error("app Load failed:", err)
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 是否开启自启动立即运行
	if *form.AutoStart {
		glogger.GLogger.Debugf("App autoStart allowed:%s-%s-%s", newUUID, form.Version, form.Name)
		if err2 := applet.StartApp(newUUID); err2 != nil {
			glogger.GLogger.Error("App autoStart failed:", err2)
		}
	}
	c.JSON(common.HTTP_OK, common.OkWithData("app create successfully"))
}

/*
*
* Update app
*
 */
func UpdateApp(c *gin.Context, ruleEngine typex.Rhilex) {
	form := AppletDto{}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	if ok, r := utils.IsValidNameLength(form.Name); !ok {
		c.JSON(common.HTTP_OK, common.Error(r))
		return
	}
	// 校验语法
	if err1 := applet.ValidateLuaSyntax([]byte(form.LuaSource)); err1 != nil {
		c.JSON(common.HTTP_OK, common.Error400(err1))
		return
	}
	mApp := model.MApplet{
		UUID:        form.UUID,
		Name:        form.Name,
		Version:     form.Version,
		AutoStart:   form.AutoStart,
		LuaSource:   form.LuaSource,
		Description: form.Description,
	}
	if err := service.UpdateApp(&mApp); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 如果内存里面有, 先把内存里的清理了
	if app := applet.GetApp(form.UUID); app != nil {
		glogger.GLogger.Debug("Already loaded, will try to stop:", form.UUID)
		// 已经启动了就不能再启动
		if app.AppState == 1 {
			applet.StopApp(form.UUID)
		}
		applet.RemoveApp(app.UUID)
	}
	// 必须先load后start
	newAPP := applet.NewApplication(mApp.UUID, mApp.Name, mApp.Version)
	newAPP.AutoStart = *mApp.AutoStart
	newAPP.Description = mApp.Description
	if err := applet.LoadApp(newAPP, mApp.LuaSource); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 自启动
	if *mApp.AutoStart {
		glogger.GLogger.Debugf("App autoStart allowed:%s-%s-%s", mApp.UUID, mApp.Version, mApp.Name)
		if err2 := applet.StartApp(mApp.UUID); err2 != nil {
			glogger.GLogger.Error("App autoStart failed:", err2)
			c.JSON(common.HTTP_OK, common.Error400(err2))
			return
		}
	} else {
		glogger.GLogger.Debugf("App autoStart not allowed:%s-%s-%s", mApp.UUID, mApp.Version, mApp.Name)
	}
	c.JSON(common.HTTP_OK, common.OkWithData("app update successfully:"+mApp.UUID))
}

/*
*
* 启动应用: 用来从数据库里面启动, 有2种情况：
* 1 停止了的, 就需要重启一下
* 2 还未被加载进来的（刚新建），先load后start
 */
func StartApp(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	// 检查数据库
	mApp, err := service.GetMAppWithUUID(uuid)
	if err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// 如果内存里面有, 判断状态
	if app := applet.GetApp(uuid); app != nil {
		glogger.GLogger.Debug("Already loaded, will try to start:", uuid)
		// 已经启动了就不能再启动
		if app.AppState == 1 {
			c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("app is running now:%s", uuid)))
		}
		if app.AppState == 0 {
			if err := applet.StartApp(uuid); err != nil {
				c.JSON(common.HTTP_OK, common.Error400(err))
			} else {
				c.JSON(common.HTTP_OK, common.OkWithData("app start successfully:"+uuid))
			}
		}
		return
	}
	// 如果内存里面没有，尝试从配置加载
	glogger.GLogger.Debug("No loaded, will try to load:", uuid)
	if err := applet.LoadApp(applet.NewApplication(
		mApp.UUID, mApp.Name, mApp.Version), mApp.LuaSource); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	glogger.GLogger.Debug("app loaded, will try to start:", uuid)
	if err := applet.StartApp(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData("app start successfully:"+uuid))
}

// 停止, 但是不删除，仅仅是把虚拟机进程给杀死
func StopApp(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	if app := applet.GetApp(uuid); app != nil {
		if app.AppState == 0 {
			c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("app is stopping now:%s", uuid)))
			return
		}
		if app.AppState == 1 {
			if err := applet.StopApp(uuid); err != nil {
				c.JSON(common.HTTP_OK, common.Error400(err))
				return
			}
			c.JSON(common.HTTP_OK, common.OkWithData("app stopped:%s"+uuid))
			return
		}
	}
	c.JSON(common.HTTP_OK, common.Error400(fmt.Errorf("app not exists:%s", uuid)))
}

// 删除
func RemoveApp(c *gin.Context, ruleEngine typex.Rhilex) {
	uuid, _ := c.GetQuery("uuid")
	// 先把正在运行的给停了
	if app := applet.GetApp(uuid); app != nil {
		app.Remove()
	}
	// 内存给清理了
	if err := applet.RemoveApp(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	// Sqlite 配置也给删了
	if err := service.DeleteApp(uuid); err != nil {
		c.JSON(common.HTTP_OK, common.Error400(err))
		return
	}
	c.JSON(common.HTTP_OK, common.OkWithData(fmt.Sprintf("remove app successfully:%s", uuid)))
}
