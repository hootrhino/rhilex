// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package mqttserver

/*
*
* 获取当前连接进来的MQTT客户端
*
 */
type _topic struct {
	Topic string
}
type Client struct {
	ID           string   `json:"id"`
	Remote       string   `json:"remote"`
	Listener     string   `json:"listener"`
	Username     string   `json:"username"`
	CleanSession bool     `json:"cleanSession"`
	Topics       []_topic `json:"topics"`
}

type PageRequest struct {
	Current int `json:"current,omitempty"`
	Size    int `json:"size,omitempty"`
}

type PageResult struct {
	Current int      `json:"current"`
	Size    int      `json:"size"`
	Total   int      `json:"total"`
	Records []Client `json:"records"`
}

// Paginate 实现分页功能
func Paginate(data []Client, pageRequest PageRequest) []Client {
	startIndex := (pageRequest.Current - 1) * pageRequest.Size
	endIndex := startIndex + pageRequest.Size
	if startIndex >= len(data) {
		return []Client{}
	}
	if endIndex > len(data) {
		endIndex = len(data)
	}
	return data[startIndex:endIndex]
}
