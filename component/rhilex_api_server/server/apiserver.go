package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	response "github.com/hootrhino/rhilex/component/rhilex_api_server/common"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/model"
	"github.com/hootrhino/rhilex/component/rhilex_api_server/service"
	"github.com/hootrhino/rhilex/core"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/ossupport"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 拼接URL
*
 */
func ContextUrl(path string) string {
	return API_V1_ROOT + path
}

const API_V1_ROOT string = "/api/v1/"

// DefaultApiServer 全局API Server
var DefaultApiServer *RhilexApiServer

/*
*
* API Server
*
 */
type RhilexApiServer struct {
	ginEngine  *gin.Engine
	ruleEngine typex.Rhilex
	config     serverConfig
}
type serverConfig struct {
	Host string `ini:"host"`
	Port int    `ini:"port"`
}

var err1crash = errors.New("http server crash, try to recovery")

/*
*
* 开启Server
*
 */
func StartRhilexApiServer(ruleEngine typex.Rhilex, port int) {
	if core.GlobalConfig.AppDebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	server := RhilexApiServer{
		ginEngine:  gin.New(),
		ruleEngine: ruleEngine,
		config:     serverConfig{Port: port},
	}
	staticFs := WWWRoot("")
	server.ginEngine.Use(static.Serve("/", staticFs))
	server.ginEngine.Use(Authorize())
	server.ginEngine.Use(CheckLicense())
	server.ginEngine.Use(Cros())
	server.ginEngine.GET("/ws", glogger.WsLogger)
	server.ginEngine.GET("/logo.svg", func(c *gin.Context) {
		c.Header("Content-Type", "image/svg+xml")
		c.FileFromFS("logo.svg", staticFs)
	})
	server.ginEngine.GET("/favicon.svg", func(c *gin.Context) {
		c.Header("Content-Type", "image/svg+xml")
		c.FileFromFS("favicon.svg", staticFs)
	})
	server.ginEngine.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		if core.GlobalConfig.AppDebugMode {
			debug.PrintStack()
		}
		c.JSON(200, response.Error500(err1crash))
	}))
	/*
	*
	* 解决浏览器刷新被重定向问题
	*
	 */
	server.ginEngine.NoRoute(func(c *gin.Context) {
		if c.ContentType() == "application/json" {
			c.Writer.WriteHeader(http.StatusNotFound)
			c.JSON(404, response.Error("No such Route:"+c.Request.URL.Path))
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Add("Accept", "text/html")
		c.Writer.Write(indexHTML)
		c.Writer.Flush()
	})
	//
	// Http server
	//
	go StartRateLimiter(typex.GCTX)
	go func(ctx context.Context, port int) {
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			glogger.GLogger.Fatalf("Http Api Server listen error: %s", err)
		}
		defer listener.Close()
		if err := server.ginEngine.RunListener(listener); err != nil {
			glogger.GLogger.Fatalf("Http Api Server listen error: %s", err)
		}
	}(typex.GCTX, server.config.Port)
	glogger.GLogger.Infof("Http Api Server listen on: %s", fmt.Sprintf("0.0.0.0:%d", port))
	DefaultApiServer = &server
}

// 即将废弃
func (s *RhilexApiServer) AddRoute(f func(c *gin.Context,
	ruleEngine typex.Rhilex)) func(*gin.Context) {
	return func(c *gin.Context) {
		f(c, s.ruleEngine)
	}
}

// New api after 0.6.4
func AddRoute(f func(c *gin.Context,
	ruleEngine typex.Rhilex)) func(*gin.Context) {
	return func(c *gin.Context) {
		f(c, DefaultApiServer.ruleEngine)
	}
}

// AddRouteV2: Only for cron，It's redundant API， will refactor in feature
func AddRouteV2(f func(*gin.Context, typex.Rhilex) (any, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := f(c, DefaultApiServer.ruleEngine)
		if err != nil {
			c.JSON(response.HTTP_OK, response.Error400(err))
		} else {
			c.JSON(response.HTTP_OK, response.OkWithData(data))
		}
	}
}

func (s *RhilexApiServer) GetGroup(name string) *gin.RouterGroup {
	return s.ginEngine.Group(name)
}

// new API
func RouteGroup(name string) *gin.RouterGroup {
	return DefaultApiServer.ginEngine.Group(name)
}
func (s *RhilexApiServer) Route() *gin.Engine {
	return s.ginEngine
}

/*
*
* 初始化网络配置,主要针对Linux，而且目前只支持了Ubuntu1804,后期分一个windows版本
*
 */
func (s *RhilexApiServer) InitializeGenericOSData() {
	initStaticModel()
}
func (s *RhilexApiServer) InitializeUnixData() {
	glogger.GLogger.Info("Initialize Unix Default Data")
}
func (s *RhilexApiServer) InitializeWindowsData() {
	glogger.GLogger.Info("Initialize Windows Default Data")
}

func (s *RhilexApiServer) InitializeEEKITData() {
	env := os.Getenv("ARCHSUPPORT")
	if env == "EEKITH3" {
		glogger.GLogger.Info("Initialize Rhilex Pi Default Data")
	}
}

/*
*
* 初始化配置
*
 */
func (s *RhilexApiServer) InitializeConfigCtl() {
	// 一组操作, 主要用来初始化 DHCP和DNS、网卡配置等
	// 1 2 3 的目的是为了每次重启的时候初始化软路由
	env := os.Getenv("ARCHSUPPORT")
	if env == "EEKITH3" {
		{
			MIproute, err := service.GetDefaultIpRoute()
			if err != nil {
				return
			}
			// 1 初始化默认路由表: ip route
			ossupport.ConfigDefaultIpTable(MIproute.Iface)
			// 2 初始化默认DHCP
			ossupport.ConfigDefaultIscServer(MIproute.Iface)
			// 3 初始化Eth1的静态IP地址
			ossupport.ConfigDefaultIscServeDhcp(ossupport.IscServerDHCPConfig{
				Iface:       MIproute.Iface,
				Ip:          MIproute.Ip,
				Network:     MIproute.Network,
				Gateway:     MIproute.Gateway,
				Netmask:     MIproute.Netmask,
				IpPoolBegin: MIproute.IpPoolBegin,
				IpPoolEnd:   MIproute.IpPoolEnd,
				IfaceFrom:   MIproute.IfaceFrom,
				IfaceTo:     MIproute.IfaceTo,
			})
		}
		{
			// 4 配置WIFI
			MWlan0, err := service.GetWlan0Config()
			if err != nil {
				return
			}
			if ossupport.WifiAlreadyConfig(MWlan0.SSID) {
				s := "nmcli connection up %s"
				shell := fmt.Sprintf(s, MWlan0.SSID)
				glogger.GLogger.Debug(shell)
				cmd := exec.Command("sh", "-c", shell)
				out, err := cmd.CombinedOutput()
				if err != nil {
					glogger.GLogger.Error(err)
					return
				}
				glogger.GLogger.Debug(string(out))
			} else {
				s := "nmcli dev wifi connect \"%s\" password \"%s\""
				shell := fmt.Sprintf(s, MWlan0.SSID, MWlan0.Password)
				glogger.GLogger.Debug(shell)
				cmd := exec.Command("sh", "-c", shell)
				out, err := cmd.CombinedOutput()
				if err != nil {
					glogger.GLogger.Error(err)
					return
				}
				glogger.GLogger.Debug(string(out))
			}

			return
		}
	}
}

/*
*
* 初始化一些静态数据
*
 */
func initStaticModel() {
	// 初始化有线网口配置
	service.InitNetWorkConfig()
	// 初始化WIFI配置
	service.InitWlanConfig()
	// 初始化默认路由, 如果没有配置会在数据库生成关于eth1的一个默认路由数据
	service.InitDefaultIpRoute()
	// 初始化硬件接口参数
	service.InitHwPortConfig()
	// 配置一个默认分组
	service.InitGenericGroup(&model.MGenericGroup{
		UUID:   "DROOT",
		Type:   "DEVICE",
		Name:   "默认分组",
		Parent: "NULL",
	})
	service.InitGenericGroup(&model.MGenericGroup{
		UUID:   "ULTROOT",
		Type:   "USER_LUA_TEMPLATE",
		Name:   "默认分组",
		Parent: "NULL",
	})
	// 初始化一个用户
	service.InitMUser(
		&model.MUser{
			Role:        "Admin",
			Username:    "rhilex",
			Password:    "25d55ad283aa400af464c76d713c07ad", // md5(12345678)
			Description: "Default RHILEX Admin User",
		},
	)
}
