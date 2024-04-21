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

# Pro1 API 配置

## WebHook
继电器打开时向网关发送数据：
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-SWITCH1-EVENT-ON-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "switch.on",
        "urls": [
            "http://192.168.1.175:6400?action=switch_on"
        ]
    }
}
```

继电器关闭时向网关发送数据：
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-SWITCH1-EVENT-OFF-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "switch.off",
        "urls": [
            "http://192.168.1.175:6400?action=switch_off"
        ]
    }
}
```
开关1打开时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT1-EVENT-ON-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "input.toggle_on",
        "urls": [
            "http://192.168.1.175:6400?action=input1_on"
        ]
    }
}
```
开关1关闭时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT1-EVENT-OFF-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "input.toggle_off",
        "urls": [
            "http://192.168.1.175:6400?action=input1_off"
        ]
    }
}
```
开关2打开时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT2-EVENT-ON-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "input.toggle_on",
        "urls": [
            "http://192.168.1.175:6400?action=input2_on"
        ]
    }
}
```
开关2关闭时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT2-EVENT-OFF-TO-RHILEX",
        "cid": 0,
        "enable": true,
        "event": "input.toggle_off",
        "urls": [
            "http://192.168.1.175:6400?action=input2_off"
        ]
    }
}
```

> 网关从URL里面取`action`，即可判断

清空所有Webhook:
- GET: http://192.168.1.106/rpc/Webhook.DeleteAll