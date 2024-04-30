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

# FRP 内网穿透
FRP（Fast Reverse Proxy）是一个高性能的反向代理应用，主要用于内网穿透。内网穿透是指在网络环境中，当某个设备或服务位于私有网络（内网）中，而外部网络（如互联网）无法直接访问时，通过特定的技术手段实现外部网络对内网资源的访问。
FRP的工作原理如下：
1. **客户端与服务端**：FRP包括客户端和服务端两部分。客户端部署在内网中的设备上，服务端部署在一个具有公网IP的机器上。
2. **代理连接**：客户端与服务器之间建立连接。客户端将内网中的服务（如Web服务、SSH服务等）通过这个连接代理到服务端。
3. **端口映射**：服务端将接收到的请求转发到内网客户端指定的端口。这样，外部网络就可以通过访问服务端的公网IP和端口来访问内网中的服务。
4. **安全与加密**：FRP支持对传输数据进行加密，确保数据传输的安全。
FRP的主要用途包括：
- **远程访问**：例如，在家访问办公室内网的电脑或服务。
- **内网服务暴露**：将内网的服务（如Web服务、数据库服务等）暴露到公网上。
- **远程调试**：开发过程中，远程调试运行在内网中的应用。
FRP因其配置简单、性能高、支持多种代理方式（如TCP、UDP、HTTP、HTTPS等）而被广泛使用。但需要注意的是，在使用FRP进行内网穿透时，应确保网络安全，避免暴露敏感信息。

## 插件功能
本插件就是提供FRP客户端，连接远程 FRP Server，实现将本地应用透传到外网。

## 环境搭建
下面给出个简单配置：
### Server
```ini
bindPort = 7000
webServer.port = 7500
webServer.user = "admin"
webServer.password = "admin"
```

### Client
配置示例：
```ini
serverAddr = "192.168.1.175"
serverPort = 7000

[[proxies]]
name = "test-tcp"
type = "tcp"
localIP = "192.168.1.175"
localPort = 2580
remotePort = 60001
```

FRP的配置文件通常分为几个部分，包括服务端配置、客户端配置和代理配置。您提供的配置片段是FRP客户端的配置，用于设置一个TCP代理。下面是对这段配置的详细解释：
1. **服务端地址和端口**：
   - `serverAddr = "192.168.1.175"`：这是FRP服务端的IP地址。FRP客户端将通过这个地址连接到服务端。
   - `serverPort = 7000`：这是FRP服务端监听的端口。客户端将使用这个端口与服务端建立连接。

2. **代理配置**：
   - `[[proxies]]`：这是一个代理配置的标识，表示接下来是一个代理的配置。
   - `name = "test-tcp"`：这是为代理设置的名称，便于识别和管理。
   - `type = "tcp"`：这表示这是一个TCP类型的代理。
   - `localIP = "192.168.1.175"`：这是本地机器（客户端所在机器）的IP地址，即内网中需要暴露的服务所在的机器。
   - `localPort = 2580`：这是本地机器上需要暴露的服务所监听的端口。
   - `remotePort = 60001`：这是FRP服务端上用于接收外部连接的端口。外部网络通过访问服务端的公网IP和这个端口来访问内网的`localPort`上的服务。

FRP客户端连接到IP地址为`192.168.1.175`，端口为`7000`的FRP服务端。然后，它将客户端所在机器（`192.168.1.175`）上的`2580`端口通过服务端的`60001`端口暴露给外部网络。这样，外部网络就可以通过访问FRP服务端的公网IP和端口`60001`来访问内网中的`2580`端口服务。

注意：*客户端和服务端版本必须一致*。

## 插件配置
本插件工作在客户端模式下，因此2我们只需要支持客户端配置即可。
```ini
[plugin.frpc]
# 启用插件
enable = true

# 服务器端地址
server_addr = "127.0.0.1"
# FRPS 监听端口
server_port = 7000

# 插件名称
name = "rhilex-web-dashboard"
# 插件类型
type = "tcp"

# 本地IP地址
local_ip = "127.0.0.1"
# 本地端口
local_port = 2580

# 服务器代理端口
remote_port = 60001

```