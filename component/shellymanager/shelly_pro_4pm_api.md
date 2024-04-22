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

# Pro1 4PM 配置
- 4 instances of Input (input:0, input:1, input:2, input:3)
- 4 instances of Switch (switch:0, switch:1, switch:2, switch:3)

## 变量
GET: `http://192.168.1.106/rpc/Webhook.ListSupported`
```json
{
    "types": {
        "input.toggle_on": {},
        "input.toggle_off": {},
        "input.button_push": {},
        "input.button_longpush": {},
        "input.button_doublepush": {},
        "input.button_triplepush": {},
        "input.analog_change": {
            "attrs": [
                {
                    "name": "percent",
                    "type": "number",
                    "desc": "New voltage in %"
                },
                {
                    "name": "xpercent",
                    "type": "number",
                    "desc": "Transformed voltage value"
                }
            ]
        },
        "input.analog_measurement": {
            "attrs": [
                {
                    "name": "percent",
                    "type": "number",
                    "desc": "New voltage in %"
                },
                {
                    "name": "xpercent",
                    "type": "number",
                    "desc": "Transformed voltage value"
                }
            ]
        },
        "switch.off": {},
        "switch.on": {}
    }
}
```

## URL token
事件也支持URL令牌替换。在调用URL之前，会解析格式为`${token}`的标记。token可以是有效的JavaScript表达式。`${token}`将被此表达式计算的结果所替换。如果计算失败，则token的内容将原样复制。在计算过程中，对象`config、status、info`以及带有属性的事件的`ev`或`event`作为条件是可用的：

- status是一个对象，它包含了由`Shelly.GetStatus`返回的整个设备状态。
- config是一个对象，它包含了由`Shelly.GetConfig`返回的整个设备配置。
- info是一个对象，它包含了由`Shelly.GetDeviceInfo`返回的设备信息。


提示：`attrs`是能带在URL里面的参数，比如：http://example.com/endpoint?percent=${event.percent}

## WebHook
继电器打开时向网关发送数据：
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-SWITCH${cid}-EVENT-ON-TO-RHILEX",
        "cid": "${cid}",
        "enable": true,
        "event": "switch.on",
        "urls": [
            "http://192.168.1.175:6400?mac=${config.sys.device.mac}&token=shelly&action=switch_on&cid=${cid}"
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
        "name":"PUSH-SWITCH${cid}-EVENT-OFF-TO-RHILEX",
        "cid": "${cid}",
        "enable": true,
        "event": "switch.off",
        "urls": [
            "http://192.168.1.175:6400?mac=${config.sys.device.mac}&token=shelly&action=switch${cid}_off&cid=${cid}"
        ]
    }
}
```
开关打开时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT${cid}-EVENT-ON-TO-RHILEX",
        "cid": "${cid}",
        "enable": true,
        "event": "input.toggle_on",
        "urls": [
            "http://192.168.1.175:6400?mac=${config.sys.device.mac}&token=shelly&action=input${cid}_on&cid=${cid}"
        ]
    }
}
```
开关关闭时向网关发送数据:
```json
{
    "id": 1,
    "method": "Webhook.Create",
    "params": {
        "name":"PUSH-INPUT${cid}-EVENT-OFF-TO-RHILEX",
        "cid": "${cid}",
        "enable": true,
        "event": "input.toggle_off",
        "urls": [
            "http://192.168.1.175:6400?mac=${config.sys.device.mac}&token=shelly&action=input${cid}_off&cid=${cid}"
        ]
    }
}
```

> 网关从URL里面取`action`，即可判断

清空所有Webhook:
- GET: http://192.168.1.106/rpc/Webhook.DeleteAll