# Copyright (C) 2024 wwhai
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

import requests
import json

url = "http://127.0.0.1:2580/api/v1/devices/create"

payload = json.dumps(
    {
        "name": "GENERIC_SNMP",
        "type": "GENERIC_SNMP",
        "gid": "DROOT",
        "schemaId": "",
        "config": {
            "commonConfig": {"autoRequest": True},
            "snmpConfig": {
                "timeout": 5,
                "frequency": 5,
                "target": "192.168.1.170",
                "port": 161,
                "transport": "udp",
                "community": "public",
                "version": 3,
            },
        },
        "description": "GENERIC_SNMP",
    }
)
headers = {"Content-Type": "application/json"}

response = requests.request("POST", url, headers=headers, data=payload)

print(response.text)
