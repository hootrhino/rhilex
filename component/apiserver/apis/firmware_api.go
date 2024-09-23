// Copyright (C) 2023 wwhai
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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// System 固件相关操作
package apis

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/hootrhino/rhilex/component/apiserver/common"
	"github.com/hootrhino/rhilex/ossupport"
	"github.com/hootrhino/rhilex/typex"
)

/*
*
* 重启固件
*
 */
func ReStartRhilex(c *gin.Context, ruleEngine typex.Rhilex) {
	if runtime.GOOS == "windows" {
		c.JSON(common.HTTP_OK, common.Error("Not support windows!"))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
	ruleEngine.Stop()
	os.Exit(0)
}

/*
*
* 重启操作系统
*
 */
func Reboot(c *gin.Context, ruleEngine typex.Rhilex) {
	if runtime.GOOS == "windows" {
		c.JSON(common.HTTP_OK, common.Error("Not support windows!"))
		return
	}
	err := ossupport.Reboot()
	if err != nil {
		c.JSON(common.HTTP_OK, common.OkWithData(err))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
}

/*
*
* 恢复出厂, 直接删除配置即可,但是现阶段暂时不实现
*
 */
func RecoverNew(c *gin.Context, ruleEngine typex.Rhilex) {
	if runtime.GOOS == "windows" {
		c.JSON(common.HTTP_OK, common.Error("Not support windows!"))
		return
	}
	c.JSON(common.HTTP_OK, common.Ok())
	ossupport.StartRecoverProcess()
}

/*
*
* 获取本地升级日志
*
 */
func GetUpGradeLog(c *gin.Context, ruleEngine typex.Rhilex) {
	byteS, _ := os.ReadFile(ossupport.UpgradeLogPath)
	c.JSON(common.HTTP_OK, common.OkWithData(string(byteS)))
}

/*
*
* 下载运行日志
*
 */
func GetRunningLog(c *gin.Context, ruleEngine typex.Rhilex) {
	c.Writer.WriteHeader(http.StatusOK)
	if RunningLogPathExists(ossupport.RunningLogPath) {
		c.FileAttachment(ossupport.RunningLogPath,
			fmt.Sprintf("running_log_%d_.txt", time.Now().UnixNano()))
	} else {
		js := `<script>alert("log file not found");window.location.href = "/";</script>`
		c.Writer.Write([]byte(js))
	}
	c.Writer.Flush()

}
func RunningLogPathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
