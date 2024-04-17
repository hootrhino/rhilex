<!--
 Copyright (C) 2024 wwhai

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

# 测试规范
go的测试规范建议是：单元测试和业务模块写在一起，但是这么写很明显不符合文件组织管理习惯，所以RHILEX得单元测试规范做个调整，不用兼容go的规范。
## 约定
- test: 目录是存放所有单元测试文件的地方
- test/apps: 是存放所有轻量应用测试脚本的地方
- test/data: 是存放所有需要使用的配置文件以及数据的地方
- test/lua: 是存放所有单元测试、集成测试用到的规则脚本的地方
- test/script: 存放Python等脚本
- test/trailer: 存放trailer的配置（现阶段这个特性暂时不上线）

## 测试规范
首先本地单元测试覆盖到，才可提交到github。后期会在github做hook，没有单元测试不允许合并。

## 单元测试
普通函数级功能直接在go的unit test框架内完成即可，比如下面这个获取IP地址的示例。
```go
func TestGetLocalIp(t *testing.T) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(ipnet.IP.String())
			}
		}
	}
}

```
## 继承测试
### 启动RHILEX实例
```go
engine := RunTestEngine()
engine.Start()
```

### 加载设备

```go
GENERIC_AIS_RECEIVER := typex.NewDevice(typex.GENERIC_AIS_RECEIVER,
	"GENERIC_AIS_RECEIVER", "GENERIC_AIS_RECEIVER", map[string]interface{}{
		"host": "0.0.0.0",
		"port": 6005,
	})
ctx, cancelF := typex.NewCCTX()
if err := engine.LoadDeviceWithCtx(GENERIC_AIS_RECEIVER, ctx, cancelF); err != nil {
	t.Fatal(err)
}
```

### 完整实例
测试AIS接收设备的生命周期:
```go
func Test_generic_ais_txrx_device(t *testing.T) {
	engine := RunTestEngine()
	engine.Start()

	hh := httpserver.NewHttpApiServer(engine)
	if err := engine.LoadPlugin("plugin.http_server", hh); err != nil {
		t.Fatal(err)
	}
	GENERIC_AIS_RECEIVER := typex.NewDevice(typex.GENERIC_AIS_RECEIVER,
		"GENERIC_AIS_RECEIVER", "GENERIC_AIS_RECEIVER", map[string]interface{}{
			"host": "0.0.0.0",
			"port": 6005,
		})
	ctx, cancelF := typex.NewCCTX()
	if err := engine.LoadDeviceWithCtx(GENERIC_AIS_RECEIVER, ctx, cancelF); err != nil {
		t.Fatal(err)
	}
	time.Sleep(25 * time.Second)
	engine.Stop()
}
```
## 对API进行测试
测试HTTP API 使用Python完成。例如下面这个新建modbus设备的示例。
```py
import requests
import json

url = "http://127.0.0.1:2580/api/v1/devices/create"

payload = json.dumps({
  "name": "Modbus设备测试",
  "type": "GENERIC_MODBUS",
  "description": "Modbus设备测试",
  "gid": "DROOT",
  "config": {
    "commonConfig": {
      "mode": "UART",
      "autoRequest": True,
      "enableOptimize": False
    },
    "portUuid": "COM12"
  },
  "schemaId": ""
})
headers = {
  'Content-Type': 'application/json'
}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
```