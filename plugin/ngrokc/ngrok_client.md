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

# Ngrok 客户端
## 关于Ngrok
Ngrok 是一个开源的网络代理工具，它允许你将本地开发环境的端口映射到互联网上，从而使远程用户能够通过浏览器访问你的本地服务。这对于测试、演示或远程调试应用程序非常有用，特别是在需要从外部访问本地网络服务时。
Ngrok 的主要特点包括：
1. **端口映射**：它允许你将本地机器上的任意端口映射到互联网上，通常是通过一个自定义的 URL。
2. **HTTP/HTTPS 隧道**：Ngrok 支持 HTTP 和 HTTPS 隧道，这意味着你可以通过安全的加密连接访问你的本地服务。
3. **远程访问**：它允许你从任何地方访问你的本地服务，无论是通过电脑、手机还是平板电脑。
4. **安全认证**：Ngrok 提供了一个安全认证系统，允许你限制对映射服务的访问。
5. **实时日志**：它提供了一个实时的日志系统，让你可以监控和调试你的服务。
6. **团队协作**：Ngrok 支持团队协作，允许团队成员共享和访问相同的映射服务。
Ngrok 通常与命令行界面一起使用，但也有一个图形用户界面。它适用于多种操作系统，包括 Windows、macOS 和 Linux。
Ngrok 是一个流行的工具，尤其是在开发和测试阶段，因为它可以简化远程访问本地服务的流程。然而，Ngrok 是一个付费服务，尽管有一个有限的免费计划，但如果你需要更高级的功能，你可能需要购买订阅。

## Ngrok插件
本插件借助Ngrok提供的免费服务，实现将本地端口映射到公网，从而实现内网透传。

## 配置
```ini
[plugin.ngrokc]
# 启用插件
enable = true
# 服务器端地址
server_endpoint = "default"
# 认证参数
auth_token = "auth_token"
# tcp | http | https
local_schema = "http"
# 本地IP地址
local_host = "127.0.0.1"
# 本地端口
local_port = 2580

```
## 指令
获取配置：
```json
{
    "uuid": "NGROKC",
    "name": "get_config",
    "args": []
}
```
启动：
```json
{
    "uuid": "NGROKC",
    "name": "start",
    "args": []
}
```
停止:
```json
{
    "uuid": "NGROKC",
    "name": "stop",
    "args": []
}
```
## 测试
```go
package main

import (
	"fmt"
	"net/http"
)

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello World, Ngrok Running</h1>")
}

func StartHttpServer() {
	http.HandleFunc("/", helloWorldHandler)
	err := http.ListenAndServe(":2589", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

```
## 参考
- https://ngrok.com/docs/getting-started/go/