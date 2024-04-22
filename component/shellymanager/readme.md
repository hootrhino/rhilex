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

# Shelly 智能家居管理网关
保加利亚一个智能家居品牌，文档在这：https://shelly-api-docs.shelly.cloud。这里主要做了一些局域网设备管理功能。

## 0.7.0 支持
第一阶段先支持Shelly的3款设备：
- ShellyPlus2PM：https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPlus2PM
- ShellyPro1：https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPro1
- ShellyPro4PM：https://shelly-api-docs.shelly.cloud/gen2/Devices/Gen2/ShellyPro4PM

## 主要功能
- 本网段扫描，然后自动加入管理组，扫描方式：局域网内 arp，shelly设备收到会回复，过滤一遍即可拿到IP和MAC，让后将其存入内存刷新列表即可。同时起一个心跳检查器，每隔5秒全表扫描一遍，请求shelly的API接口，来确认是否上下线等。
- 同时shelly设备内有个webhook可以用来配置成通知功能。
- 一键配置WebHook
- 一键清空WebHook配置