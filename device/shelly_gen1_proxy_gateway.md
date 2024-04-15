# Shelly 1代设备本地管理管理
Shelly 设备支持MQTT，本插件的原理就是起一个本地版MQTT Server，代理设备的请求，同时可以给设备下发配置等。

## 文档
- https://shelly-api-docs.shelly.cloud/gen1

## 计划
第一阶段先支持Shelly的3款设备：
- ShellyPlus2PM: https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPlus2PM
- ShellyPro1: https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPro1
- ShellyPro4PM: https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPro4PM

## 配置
```json
{
    "name": "SHELLY_GEN1_PROXY_SERVER",
    "type": "SHELLY_GEN1_PROXY_SERVER",
    "schemaId": "",
    "gid": "DROOT",
    "config": {
            "networkCidr": "192.168.1.0/24",
            "autoRequest": true,
            "timeout": 3000,
            "frequency": 5000
    },
    "description": "SHELLY_GEN1_PROXY_SERVER"
}
```