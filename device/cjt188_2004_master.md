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
# CJT188数据采集网关

## 配置
```json
{
    "name": "CJT1882004_MASTER",
    "type": "CJT1882004_MASTER",
    "description": "CJT1882004_MASTER",
    "gid": "DROOT",
    "config": {
        "commonConfig": {
            "mode": "UART",
            "autoRequest": true,
            "batchRequest": false
        },
        "hostConfig": {
            "host": "192.168.1.100",
            "port": 6000,
            "timeout": 5000
        },
        "uartConfig": {
            "uart": "COM9",
            "baudRate": 2400,
            "dataBits": 8,
            "stopBits": 1,
            "parity": "E"
        }
    }
}
```
