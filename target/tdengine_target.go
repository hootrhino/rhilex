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

package target

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hootrhino/rhilex/common"
	"github.com/hootrhino/rhilex/glogger"
	"github.com/hootrhino/rhilex/typex"
	"github.com/hootrhino/rhilex/utils"
)

/*
*
* TDengine 的资源输出支持,当前暂时支持HTTP接口的形式，后续逐步会增加UDP、TCP模式
*
 */

type tdEngineTarget struct {
	typex.XStatus
	client     http.Client
	mainConfig common.TDEngineConfig
	status     typex.SourceState
}
type tdHttpResult struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Desc   string `json:"desc"`
}

func NewTdEngineTarget(e typex.Rhilex) typex.XTarget {
	td := tdEngineTarget{
		client:     http.Client{},
		mainConfig: common.TDEngineConfig{},
	}
	td.RuleEngine = e
	td.status = typex.SOURCE_DOWN
	return &td

}
func (td *tdEngineTarget) test() bool {
	if err := execQuery(td.client,
		td.mainConfig.Username,
		td.mainConfig.Password,
		"SELECT CLIENT_VERSION();",
		td.url()); err != nil {
		glogger.GLogger.Error(err)
		return false
	}
	return true
}
func (td *tdEngineTarget) url() string {
	return fmt.Sprintf("http://%s:%v/rest/sql/%s",
		td.mainConfig.Fqdn, td.mainConfig.Port, td.mainConfig.DbName)
}

//
// 注册InEndID到资源
//

func (td *tdEngineTarget) Init(outEndId string, configMap map[string]interface{}) error {
	td.PointId = outEndId

	if err := utils.BindSourceConfig(configMap, &td.mainConfig); err != nil {
		return err
	}
	if td.test() {
		return nil
	}
	return errors.New("tdengine connect error")
}

// 启动资源
func (td *tdEngineTarget) Start(cctx typex.CCTX) error {
	td.Ctx = cctx.Ctx
	td.CancelCTX = cctx.CancelCTX
	//
	td.status = typex.SOURCE_UP
	return nil
}

// 数据模型, 用来描述该资源支持的数据, 对应的是云平台的物模型
func (td *tdEngineTarget) DataModels() []typex.XDataModel {
	return []typex.XDataModel{}
}

// 获取资源状态
func (td *tdEngineTarget) Status() typex.SourceState {
	if td.test() {
		return typex.SOURCE_UP
	}
	return td.status
}

// 获取资源绑定的的详情
func (td *tdEngineTarget) Details() *typex.OutEnd {
	return td.RuleEngine.GetOutEnd(td.PointId)

}

// 停止资源, 用来释放资源
func (td *tdEngineTarget) Stop() {
	td.status = typex.SOURCE_DOWN
	if td.CancelCTX != nil {
		td.CancelCTX()
	}
}

func post(client http.Client,
	username string,
	password string,
	sql string,
	url string) (string, error) {
	body := strings.NewReader(sql)
	request, _ := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", "text/plain")
	request.SetBasicAuth(username, password)
	response, err2 := client.Do(request)
	if err2 != nil {
		return "", err2
	}
	if response.StatusCode != 200 {
		bytes0, err3 := io.ReadAll(response.Body)
		if err3 != nil {
			return "", err3
		}
		return "", fmt.Errorf("Error:%v", string(bytes0))
	}
	bytes1, err3 := io.ReadAll(response.Body)
	if err3 != nil {
		return "", err3
	}
	return string(bytes1), nil
}

/*
*
* 执行TdEngine的查询
*
 */
func execQuery(client http.Client, username string, password string, sql string, url string) error {
	var r tdHttpResult
	// {"status":"error","code":534,"desc":"Syntax error in SQL"}
	body, err1 := post(client, username, password, sql, url)
	if err1 != nil {
		return err1
	}
	err2 := utils.TransformConfig([]byte(body), &r)
	if err2 != nil {
		return err2
	}
	if r.Status == "error" {
		return fmt.Errorf("code;%v, error:%s", r.Code, r.Desc)
	}
	return nil
}

/*
* SQL: INSERT INTO meter VALUES (NOW, %v, %v);
* 数据到达后写入Tdengine, 这里对数据有严格约束，必须是以,分割的字符串
* 比如: 10.22,220.12,123,......
*
 */
func (td *tdEngineTarget) To(data interface{}) (interface{}, error) {
	switch s := data.(type) {
	case string:
		{
			return execQuery(td.client, td.mainConfig.Username,
				td.mainConfig.Password, s, td.url()), nil
		}
	}
	return nil, nil
}
