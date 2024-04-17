## SNMP协议
SNMP（Simple Network Management Protocol）是一种用于网络管理的协议。它被广泛用于监视和管理网络中的设备、系统和应用程序。SNMP允许网络管理员通过网络收集和组织设备的管理信息，以便实时监控网络的状态、性能和健康状况。

SNMP基于客户端-服务器模型，其中有两个主要组件：

1. 管理器（Manager）：管理器是网络管理员或管理系统的一部分，用于监视和控制网络设备。它通过SNMP协议发送请求并接收响应，从被管理设备中获取信息并采取必要的操作。

2. 代理（Agent）：代理是安装在被管理设备上的软件模块，负责收集和维护设备的管理信息，并响应来自管理器的请求。代理将设备的状态、配置和性能信息以适当的格式暴露给管理器。

SNMP使用一组标准的管理信息库（Management Information Base，MIB）来定义设备和系统的管理信息。MIB是一个层次化的数据库，包含了设备的各种参数、统计信息和配置设置。管理器可以通过SNMP协议向代理发送请求，例如获取特定参数的值、设置参数的值、触发操作等。

SNMP协议支持各种版本，其中最常用的是SNMPv1、SNMPv2c和SNMPv3。每个版本都具有不同的功能和安全性特性，以适应不同的网络管理需求。
## 设备配置
```json
{
    "name": "GENERIC_SNMP",
    "type": "GENERIC_SNMP",
    "gid": "DROOT",
    "schemaId": "",
    "config": {
        "commonConfig": {
            "autoRequest": true
        },
        "snmpConfig": {
            "timeout": 5,
            "frequency": 5,
            "target": "192.168.1.170",
            "port": 161,
            "transport": "udp",
            "community": "public",
            "version": 3
        }
    },
    "description": "GENERIC_SNMP"
}
```
## 数据示例
```json
[
    {
        "oid": ".1.3.6.1.2.1.1.7.0",
        "tag": "sysServices",
        "alias": "sysServices",
        "value": 76
    },
    {
        "oid": ".1.3.6.1.2.1.1.1.0",
        "tag": "SystemDescription",
        "alias": "SystemDescription",
        "value": "Hardware: Intel64"
    },
    {
        "oid": ".1.3.6.1.2.1.1.5.0",
        "tag": "PCName",
        "alias": "PCName",
        "value": "DESKTOP-4LMLO5C"
    }
]
```
## 数据解析示例
```lua
function (args)
    Debug(args)
end
```